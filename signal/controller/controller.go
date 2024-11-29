// Package controller handles HTTP logic.
package controller

import (
	"fmt"
	"log"
	"net/http"
	"pdn/broker"
	"pdn/pkg/socket"
	"pdn/types/api/request"
	"pdn/types/api/response"
	"pdn/types/message"
)

const (
	PUSH = "PUSH"
	PULL = "PULL"
)

// SocketController handles HTTP requests.
type SocketController struct {
	broker *broker.Broker
}

// New creates a new instance of SocketController.
func New(b *broker.Broker) *SocketController {
	return &SocketController{
		broker: b,
	}
}

// Process handles HTTP requests.
func (c *SocketController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s, err := socket.New(w, r)
	if err != nil {
		log.Printf("Failed to create WebSocket: %v", err)
		return
	}
	defer func() {
		if err := s.Close(); err != nil {
			log.Printf("Failed to close WebSocket: %v", err)
		}
	}()

	channelID, userID, err := c.authenticate(s)
	if err != nil {
		return
	}

	// 01. Register client to broker
	if err := c.broker.Register(broker.CLIENT, broker.FROM(channelID+userID)); err != nil {
		log.Printf("Failed to publish to broker: %v", err)
		return
	}
	defer func() {
		if err := c.broker.Unregister(broker.CLIENT, broker.FROM(channelID+userID)); err != nil {
			log.Printf("Failed to unregister broker: %v", err)
		}
	}()

	// 02. subscribe itself
	go c.response(s, channelID, userID)

	// 03. send message to broker
	if err := c.publish(s, channelID, userID); err != nil {
		return
	}
}

func (c *SocketController) authenticate(s socket.Socket) (string, string, error) {
	var req request.Activate

	if err := s.ReadJSON(&req); err != nil {
		log.Printf("Failed to read activation request: %v", err)
		return "", "", err
	}

	res := response.Activate{
		RequestID:  req.RequestID,
		StatusCode: 200,
		Message:    "Connection established",
	}

	if err := s.WriteJSON(res); err != nil {
		log.Printf("Failed to write response: %v", err)
		return "", "", err
	}

	return req.ChannelID, req.UserID, nil
}

func (c *SocketController) response(s socket.Socket, channelID, userID string) {
	ch, err := c.broker.Subscribe(broker.CLIENT, broker.FROM(channelID+userID))
	if err != nil {
		log.Printf("Failed to subscribe to broker: %v", err)
		return
	}
	for {
		msg, ok := <-ch
		if !ok {
			return
		}
		if err := s.WriteJSON(msg); err != nil {
			log.Printf("Failed to write message: %v", err)
			return
		}
	}
}

func (c *SocketController) publish(s socket.Socket, channelID, userID string) error {
	for {
		// 01. read message from client
		var req request.Signal
		if err := s.ReadJSON(req); err != nil {
			log.Printf("Failed to read message: %v", err)
			return err
		}

		// 02. publish message to broker
		switch req.Type {
		case PUSH:
			if err := c.broker.Publish(broker.PUSH, broker.CONTROLLER, message.Push{
				ChannelID: channelID,
				UserID:    userID,
				SDP:       req.SDP,
			}); err != nil {
				log.Printf("Failed to publish to broker: %v", err)
			}
		case PULL:
			if err := c.broker.Publish(broker.PULL, broker.CONTROLLER, message.Pull{
				ChannelID: channelID,
				UserID:    userID,
				SDP:       req.SDP,
			}); err != nil {
				log.Printf("Failed to publish to broker: %v", err)
			}
		default:
			return fmt.Errorf("invalid type: %s", req.Type)
		}

	}
}
