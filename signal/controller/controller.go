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
	"pdn/metric"
	"pdn/types/client/request"
	"pdn/types/client/response"
	"pdn/types/message"
)

// Controller handles HTTP requests.
type Controller struct {
	broker   *broker.Broker
	database database.Database
	metric   *metric.Metrics
}

// New creates a new instance of Controller.
func New(b *broker.Broker, db database.Database, m *metric.Metrics) *Controller {
	return &Controller{
		broker:   b,
		database: db,
		metric:   m,
	}
}

// Process handles HTTP requests.
func (c *Controller) Process(conn *websocket.Conn) error {
	c.metric.IncrementWebSocketConnections()
	defer c.metric.DecrementWebSocketConnections()

	c.metric.IncrementClientConnectionAttempts()

	// 01. Build the context for control response goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
	}()

	// 02. Authenticate the connection
	channelID, userID, err := c.authenticate(conn)
	if err != nil {
		c.metric.IncrementClientConnectionFailures()
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	if err := c.broker.Publish(broker.Client, broker.ACTIVATE, message.Activate{
		ChannelID: channelID,
		ClientID:  userID,
	}); err != nil {
		c.metric.IncrementClientConnectionFailures()
		return fmt.Errorf("failed to publish connected message: %w", err)
	}
	defer func() {
		if err := c.broker.Publish(broker.Client, broker.DEACTIVATE, message.Deactivate{
			ChannelID: channelID,
			ClientID:  userID,
		}); err != nil {
			log.Printf("failed to publish left message: %v", err)
		}
	}()

	c.metric.IncrementClientConnectionSuccesses()

	go c.sendResponse(ctx, conn, channelID, userID)

	if err := c.receiveRequest(conn, channelID, userID); err != nil {
		return fmt.Errorf("failed to receive request: %w", err)
	}
	return nil
}

// authenticate authenticates the connection.
func (c *Controller) authenticate(conn *websocket.Conn) (string, string, error) {
	// 01. Parse the request from the client
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

	// 02. Authenticate the channel
	channelInfo, err := c.database.FindChannelInfoByID(payload.ChannelID)
	if err != nil {
		return "", "", fmt.Errorf("failed to find channel info: %w", err)
	}
	if !channelInfo.Authenticate(payload.ChannelKey) {
		return "", "", fmt.Errorf("invalid key: %s", payload.ChannelKey)
	}

	res := response.Activate{
		Type:    response.ACTIVATE,
		Message: "FetchFromPeer established",
	}

	if err := conn.WriteJSON(res); err != nil {
		return "", "", fmt.Errorf("failed to send activation response: %w", err)
	}

	return payload.ChannelID, payload.ClientID, nil
}

// sendResponse sends response to the client.
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

// receiveRequest receives request from the websocket and call handleRequest.
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

// handleRequest parse the request type and call the corresponding handler function
func (c *Controller) handleRequest(req request.Common, channelID, userID string) error {
	var err error
	switch req.Type {
	case request.PUSH:
		err = c.handlePush(req, channelID, userID)
	case request.PULL:
		err = c.handlePull(req, channelID, userID)
	case request.FORWARD:
		err = c.handleForward(req, channelID, userID)
	case request.SIGNAL:
		err = c.handleSignal(req, channelID, userID)
	case request.CONNECTED:
		err = c.handleConnected(req, channelID, userID)
	case request.DISCONNECTED:
		err = c.handleDisconnected(req, channelID, userID)
	case request.FAILED:
		err = c.handleFailed(req, channelID, userID)
	default:
		err = fmt.Errorf("invalid request type: %s", req.Type)
	}
	return err
}

// handlePush handles the push event. push event means that a client will push stream to media server.
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
	if err := c.broker.Publish(broker.Client, broker.PUSH, msg); err != nil {
		return fmt.Errorf("failed to publish push message: %w", err)
	}
	return nil
}

// handlePull handles the pull event. pull event means that a client will pull stream from anyone.
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
	if err := c.broker.Publish(broker.Client, broker.PULL, msg); err != nil {
		return fmt.Errorf("failed to publish pull message: %w", err)
	}
	return nil
}

// handleSignal handles the exchange event. exchange event means that a client will exchange SDP or
// candidate with another client.
func (c *Controller) handleSignal(req request.Common, channelID, userID string) error {
	var payload request.Signal
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal signal payload: %w", err)
	}
	connInfo, err := c.database.FindConnectionInfoByID(payload.ConnectionID)
	if err != nil {
		return fmt.Errorf("failed to find connection info: %w", err)
	}
	if !connInfo.Authorize(channelID, userID) {
		return fmt.Errorf("unauthorized connection exchange: %s", payload.ConnectionID)
	}

	counterpart := connInfo.GetCounterpart(userID)

	msg := response.Signal{
		Type:         response.SIGNAL,
		ConnectionID: payload.ConnectionID,
		SignalType:   payload.SignalType,
		SignalData:   payload.SignalData,
	}
	if err := c.broker.Publish(broker.ClientSocket, broker.Detail(channelID+counterpart), msg); err != nil {
		return fmt.Errorf("failed to publish exchange message: %w", err)
	}
	return nil
}

// handleForward handles the forward event. forward event means that a client requests
func (c *Controller) handleForward(req request.Common, channelID, userID string) error {
	var payload request.Forward
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal exchange payload: %w", err)
	}
	connInfo, err := c.database.FindConnectionInfoByID(payload.ConnectionID)
	if err != nil {
		return fmt.Errorf("failed to find connection info: %w", err)
	}
	if !connInfo.Authorize(channelID, userID) {
		return fmt.Errorf("unauthorized connection exchange: %s", payload.ConnectionID)
	}

	counterpart := connInfo.GetCounterpart(userID)

	msg := response.Forward{
		Type:         response.FORWARD,
		ConnectionID: payload.ConnectionID,
		SDP:          payload.SDP,
	}
	if err := c.broker.Publish(broker.ClientSocket, broker.Detail(channelID+counterpart), msg); err != nil {
		return fmt.Errorf("failed to publish exchange message: %w", err)
	}
	return nil
}

// handleConnected handles the succeed event. succeed event means that a client has successfully
// completed the connection with another client.
func (c *Controller) handleConnected(req request.Common, channelID, userID string) error {
	var payload request.Connected
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal succeed payload: %w", err)
	}
	connInfo, err := c.database.FindConnectionInfoByID(payload.ConnectionID)
	if err != nil {
		return fmt.Errorf("failed to find connection info: %w", err)
	}
	if !connInfo.Authorize(channelID, userID) {
		return fmt.Errorf("unauthorized connection exchange: %s", payload.ConnectionID)
	}

	if err := c.broker.Publish(broker.Peer, broker.CONNECTED, message.Connected{
		ConnectionID: payload.ConnectionID,
	}); err != nil {
		return fmt.Errorf("failed to publish succeed message: %w", err)
	}
	return nil
}

// handleFailed handles the failed event. failed event means that a client has failed to
// complete the connection with another client.
func (c *Controller) handleFailed(req request.Common, channelID, userID string) error {
	var payload request.Failed
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal failed payload: %w", err)
	}
	connInfo, err := c.database.FindConnectionInfoByID(payload.ConnectionID)
	if err != nil {
		return fmt.Errorf("failed to find connection info: %w", err)
	}
	if !connInfo.Authorize(channelID, userID) {
		return fmt.Errorf("unauthorized connection exchange: %s", payload.ConnectionID)
	}

	if err := c.broker.Publish(broker.Peer, broker.FAILED, message.Failed{
		ConnectionID: payload.ConnectionID,
	}); err != nil {
		return fmt.Errorf("failed to publish failed message: %w", err)
	}
	return nil
}

// handleDisconnected handles the closed event. closed event means that a client has closed the connection.
func (c *Controller) handleDisconnected(req request.Common, channelID, userID string) error {
	var payload request.Disconnected
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal closed payload: %w", err)
	}
	connInfo, err := c.database.FindConnectionInfoByID(payload.ConnectionID)
	if err != nil {
		return fmt.Errorf("failed to find connection info: %w", err)
	}
	if !connInfo.Authorize(channelID, userID) {
		return fmt.Errorf("unauthorized connection exchange: %s", payload.ConnectionID)
	}

	if err := c.broker.Publish(broker.Peer, broker.DISCONNECTED, message.Disconnected{
		ConnectionID: payload.ConnectionID,
	}); err != nil {
		return fmt.Errorf("failed to publish closed message: %w", err)
	}
	return nil
}
