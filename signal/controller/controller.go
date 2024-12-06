// Package controller handles HTTP logic.
package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"pdn/broker"
	"pdn/database"
	"pdn/types/client/request"
	"pdn/types/client/response"
	"pdn/types/message"
)

// Controller handles HTTP requests.
type Controller struct {
	broker   *broker.Broker
	database database.Database
}

// New creates a new instance of Controller.
func New(b *broker.Broker, db database.Database) *Controller {
	return &Controller{
		broker:   b,
		database: db,
	}
}

// Process handles HTTP requests.
func (c *Controller) Process(conn *websocket.Conn) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
	}()
	channelID, userID, err := c.authenticate(conn)
	defer func() {
		if err := c.database.DeleteClientInfoByID(channelID, userID); err != nil {
			log.Printf("failed to delete user info: %v", err)
		}
	}()
	if err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}
	// 02. subscribe itself
	go c.sendResponse(ctx, conn, channelID, userID)

	// 03. sendResponse message to broker
	if err := c.receiveRequest(conn, channelID, userID); err != nil {
		return fmt.Errorf("failed to receive request: %w", err)
	}
	return nil
}

// authenticate authenticates the connection.
func (c *Controller) authenticate(conn *websocket.Conn) (string, string, error) {
	var req request.Common
	if err := conn.ReadJSON(&req); err != nil {
		return "", "", fmt.Errorf("failed to read authentication message: %w", err)
	}
	if req.Type != request.ACTIVATE {
		return "", "", fmt.Errorf("expected type '%s', got '%s'", request.ACTIVATE, req.Type)
	}
	var payload request.Activate
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		return "", "", fmt.Errorf("failed to unmarshal activation payload: %w", err)
	}

	channelInfo, err := c.database.FindChannelInfoByID(payload.ChannelID)
	if err != nil {
		return "", "", fmt.Errorf("failed to find channel info: %w", err)
	}
	if channelInfo.Key != payload.ChannelKey {
		return "", "", fmt.Errorf("invalid key: %s", payload.ChannelKey)
	}

	if err := c.database.CreateClientInfo(payload.ChannelID, payload.ClientID); err != nil {
		return "", "", fmt.Errorf("failed to create user info: %w", err)
	}

	res := response.Activate{
		Type:    response.ACTIVATE,
		Message: "Connection established",
	}

	if err := conn.WriteJSON(res); err != nil {
		return "", "", fmt.Errorf("failed to send activation response: %w", err)
	}

	return payload.ChannelID, payload.ClientID, nil
}

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

func (c *Controller) handleRequest(req request.Common, channelID, userID string) error {
	var err error
	switch req.Type {
	case request.PUSH:
		err = c.handlePush(req, channelID, userID)
	case request.PULL:
		err = c.handlePull(req, channelID, userID)
	default:
		err = fmt.Errorf("invalid request type: %s", req.Type)
	}
	return err
}

func (c *Controller) handlePush(req request.Common, channelID, userID string) error {
	var payload request.Push
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal push payload: %w", err)
	}

	msg := message.Push{
		ConnectionID: payload.ConnectionID,
		ChannelID:    channelID,
		ClientID:     userID,
		SDP:          payload.SDP,
	}
	if err := c.broker.Publish(broker.ClientMessage, broker.PUSH, msg); err != nil {
		return fmt.Errorf("failed to publish push message: %w", err)
	}
	return nil
}

func (c *Controller) handlePull(req request.Common, channelID, userID string) error {
	var payload request.Pull
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal pull payload: %w", err)
	}

	msg := message.Pull{
		ConnectionID: payload.ConnectionID,
		ChannelID:    channelID,
		ClientID:     userID,
		SDP:          payload.SDP,
	}
	if err := c.broker.Publish(broker.ClientMessage, broker.PULL, msg); err != nil {
		return fmt.Errorf("failed to publish pull message: %w", err)
	}
	return nil
}
