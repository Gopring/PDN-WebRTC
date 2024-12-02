// Package controller handles HTTP logic.
package controller

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"pdn/broker"
	"pdn/types/api/request"
	"pdn/types/api/response"
	"pdn/types/message"
)

const (
	PUSH = "PUSH"
	PULL = "PULL"
)

// Controller handles HTTP requests.
type Controller struct {
	broker *broker.Broker
}

// New creates a new instance of Controller.
func New(b *broker.Broker) *Controller {
	return &Controller{
		broker: b,
	}
}

// Process handles HTTP requests.
func (c *Controller) Process(socket *websocket.Conn) {
	channelID, userID, err := c.authenticate(socket)
	if err != nil {
		return
	}

	// 02. subscribe itself
	go c.sendResponse(socket, channelID, userID)

	// 03. sendResponse message to broker
	if err := c.receiveRequest(socket, channelID, userID); err != nil {
		return
	}
}

func (c *Controller) authenticate(s *websocket.Conn) (string, string, error) {
	var req request.Activate

	if err := s.ReadJSON(&req); err != nil {
		log.Printf("Failed to read activation receiveRequest: %v", err)
		return "", "", err
	}

	res := response.Activate{
		RequestID:  req.RequestID,
		StatusCode: 200,
		Message:    "Connection established",
	}

	if err := s.WriteJSON(res); err != nil {
		log.Printf("Failed to sendResponse sendResponse: %v", err)
		return "", "", err
	}

	return req.ChannelID, req.UserID, nil
}

func (c *Controller) receiveRequest(s *websocket.Conn, channelID, userID string) error {
	for {
		// 01. read message from client
		var req request.Signal
		if err := s.ReadJSON(req); err != nil {
			log.Printf("Failed to read message: %v", err)
			return err
		}

		// 02. receiveRequest message to broker
		switch req.Type {
		case PUSH:
			if err := c.broker.Publish(broker.ClientMessage, broker.PUSH, message.Push{
				ChannelID: channelID,
				UserID:    userID,
				SDP:       req.SDP,
			}); err != nil {
				log.Printf("Failed to receiveRequest to broker: %v", err)
			}
		case PULL:
			if err := c.broker.Publish(broker.ClientMessage, broker.PULL, message.Pull{
				ChannelID: channelID,
				UserID:    userID,
				SDP:       req.SDP,
			}); err != nil {
				log.Printf("Failed to receiveRequest to broker: %v", err)
			}
		default:
			return fmt.Errorf("invalid type: %s", req.Type)
		}

	}
}

func (c *Controller) sendResponse(s *websocket.Conn, channelID, userID string) {
	ch := c.broker.Subscribe(broker.ClientSocket, broker.Detail(channelID+userID))
	for {
		msg, ok := <-ch
		if !ok {
			return
		}
		if err := s.WriteJSON(msg); err != nil {
			log.Printf("Failed to sendResponse message: %v", err)
			return
		}
	}
}
