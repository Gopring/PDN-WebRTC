// Package controller handles HTTP logic.
package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"pdn/broker"
	"pdn/types/client/request"
	"pdn/types/client/response"
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
func (c *Controller) Process(socket *websocket.Conn) error {
	channelID, userID, err := c.authenticate(socket)
	if err != nil {
		return err
	}

	// 02. subscribe itself
	go c.sendResponse(socket, channelID, userID)

	// 03. sendResponse message to broker
	if err := c.receiveRequest(socket, channelID, userID); err != nil {
		return err
	}
	return nil
}

func (c *Controller) authenticate(s *websocket.Conn) (string, string, error) {
	var req request.Activate

	if err := s.ReadJSON(&req); err != nil {
		return "", "", err
	}

	if err := s.WriteJSON(response.Activate{
		RequestID:  req.RequestID,
		StatusCode: 200,
		Message:    "Connection established",
	}); err != nil {
		return "", "", err
	}

	return req.ChannelID, req.UserID, nil
}

func (c *Controller) receiveRequest(s *websocket.Conn, channelID, userID string) error {
	for {
		var req request.Common
		if err := s.ReadJSON(req); err != nil {
			return err
		}
		switch req.Type {
		case PUSH:
			var payload request.Push
			if err := json.Unmarshal(req.Payload, &payload); err != nil {
				return err
			}
			if err := c.broker.Publish(broker.ClientMessage, broker.PUSH, message.Push{
				Common:    message.Common{RequestID: req.RequestID},
				ChannelID: channelID,
				UserID:    userID,
				SDP:       payload.SDP,
			}); err != nil {
				return err
			}
		case PULL:
			var payload request.Pull
			if err := json.Unmarshal(req.Payload, &payload); err != nil {
				return err
			}
			if err := c.broker.Publish(broker.ClientMessage, broker.PULL, message.Pull{
				Common:    message.Common{RequestID: req.RequestID},
				ChannelID: channelID,
				UserID:    userID,
				SDP:       payload.SDP,
			}); err != nil {
				return err
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
