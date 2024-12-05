// Package controller handles HTTP logic.
package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"pdn/broker"
	"pdn/metric"
	"pdn/types/client/request"
	"pdn/types/client/response"
	"pdn/types/message"
)

const (
	activate = "ACTIVATE"
	push     = "PUSH"
	pull     = "PULL"
)

// Controller handles HTTP requests.
type Controller struct {
	broker  *broker.Broker
	metrics *metric.Metrics
}

// New creates a new instance of Controller.
func New(b *broker.Broker, m *metric.Metrics) *Controller {
	return &Controller{
		broker:  b,
		metrics: m,
	}
}

// Process handles HTTP requests.
func (c *Controller) Process(conn *websocket.Conn) error {
	c.metrics.IncrementClientConnectionAttempts()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	channelID, userID, err := c.authenticate(conn)
	if err != nil {
		c.metrics.IncrementClientConnectionFailures()

		return fmt.Errorf("failed to authenticate: %w", err)
	}
	c.metrics.IncrementClientConnectionSuccesses()

	go c.sendResponse(ctx, conn, channelID, userID)

	if err := c.receiveRequest(conn, channelID, userID); err != nil {
		return fmt.Errorf("failed to receive request: %w", err)
	}
	return nil
}

// authenticate performs authentication for a WebSocket connection.
func (c *Controller) authenticate(conn *websocket.Conn) (string, string, error) {
	var req request.Common
	if err := conn.ReadJSON(&req); err != nil {
		return "", "", fmt.Errorf("failed to read authentication message: %w", err)
	}
	if req.Type != activate {
		return "", "", fmt.Errorf("expected type '%s', got '%s'", activate, req.Type)
	}
	var payload request.Activate
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		return "", "", fmt.Errorf("failed to unmarshal activation payload: %w", err)
	}

	res := response.Activate{
		RequestID:  req.RequestID,
		StatusCode: 200,
		Message:    "Connection established",
	}

	if err := conn.WriteJSON(res); err != nil {
		return "", "", fmt.Errorf("failed to send activation response: %w", err)
	}

	return payload.ChannelID, payload.UserID, nil
}

// sendResponse listens for messages from the broker and forwards them to the WebSocket client.
func (c *Controller) sendResponse(ctx context.Context, conn *websocket.Conn, channelID, userID string) {
	detail := broker.Detail(channelID + userID)
	sub := c.broker.Subscribe(broker.ClientSocket, detail)
	defer func() {
		if err := c.broker.Unsubscribe(broker.ClientSocket, detail, sub); err != nil {
			log.Printf("Error occurs in unsubscribe: %v", err)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-sub.Receive():
			if err := conn.WriteJSON(msg); err != nil {
				log.Printf("Failed to send response: %v", err)
				return
			}
		}
	}
}

// receiveRequest reads and handles incoming requests from the WebSocket client.
func (c *Controller) receiveRequest(conn *websocket.Conn, channelID, userID string) error {
	for {
		var req request.Common
		if err := conn.ReadJSON(&req); err != nil {
			return fmt.Errorf("failed to parse common message: %v", err)
		}
		if err := c.handleRequest(req, channelID, userID); err != nil {
			log.Printf("Error handling request: %v", err)
			continue
		}
	}
}

// handleRequest routes a request to the appropriate handler based on its type.
func (c *Controller) handleRequest(req request.Common, channelID, userID string) error {
	var err error
	switch req.Type {
	case push:
		err = c.handlePush(req, channelID, userID)
	case pull:
		err = c.handlePull(req, channelID, userID)
	default:
		err = fmt.Errorf("invalid request type: %s", req.Type)
	}
	return err
}

// handlePush processes a "PUSH" request from the WebSocket client.
func (c *Controller) handlePush(req request.Common, channelID, userID string) error {
	var payload request.Push
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal push payload: %w", err)
	}

	msg := message.Push{
		Common:    message.Common{RequestID: req.RequestID},
		ChannelID: channelID,
		UserID:    userID,
		SDP:       payload.SDP,
	}
	if err := c.broker.Publish(broker.ClientMessage, broker.PUSH, msg); err != nil {
		return fmt.Errorf("failed to publish push message: %w", err)
	}
	return nil
}

// handlePull processes a "PULL" request from the WebSocket client.
func (c *Controller) handlePull(req request.Common, channelID, userID string) error {
	var payload request.Pull
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal pull payload: %w", err)
	}

	msg := message.Pull{
		Common:    message.Common{RequestID: req.RequestID},
		ChannelID: channelID,
		UserID:    userID,
		SDP:       payload.SDP,
	}
	if err := c.broker.Publish(broker.ClientMessage, broker.PULL, msg); err != nil {
		return fmt.Errorf("failed to publish pull message: %w", err)
	}
	return nil
}
