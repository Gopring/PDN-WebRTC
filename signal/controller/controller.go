// Package controller handles HTTP logic.
package controller

import (
	"fmt"
	"log"
	"pdn/pkg/socket"
	"pdn/signal/coordinator"
	"pdn/types/api/request"
)

// Request types for signaling in the socket communication.
const (
	PUSH      = "PUSH"
	PULL      = "PULL"
	FORWARD   = "FORWARD"
	FETCH     = "FETCH"
	ARRANGE   = "ARRANGE"
	RECONNECT = "RECONNECT"
)

// SocketController handles HTTP requests.
type SocketController struct {
	coordinator coordinator.Coordinator
	debug       bool
}

// New creates a new instance of SocketController.
func New(c coordinator.Coordinator, isDebug bool) *SocketController {
	return &SocketController{
		coordinator: c,
		debug:       isDebug,
	}
}

// Process handles HTTP requests.
func (c *SocketController) Process(s socket.Socket) error {
	var req request.Activate
	if err := s.Read(&req); err != nil {
		log.Printf("Failed to read activation request: %v", err)
		return err
	}

	if err := c.coordinator.Activate(req.ChannelID, req.UserID, s); err != nil {
		log.Printf("Failed to activate coordinator (ChannelID: %s, UserID: %s): %v", req.ChannelID, req.UserID, err)
		return err
	}
	if err := s.Write("Activated"); err != nil {
		log.Printf("Failed to write response: %v", err)
		return err
	}

	defer func() {
		if err := c.coordinator.Remove(req.ChannelID, req.UserID); err != nil {
			log.Printf("Failed to remove coordinator (ChannelID: %s, UserID: %s): %v", req.ChannelID, req.UserID, err)
		}
	}()

	log.Printf("Connection established (ChannelID: %s, UserID: %s)", req.ChannelID, req.UserID)
	if err := c.handleConnection(s, req.ChannelID, req.UserID); err != nil {
		return fmt.Errorf("connection handling error (ChannelID: %s, UserID: %s): %v", req.ChannelID, req.UserID, err)
	}
	return nil
}

// handleConnection processes incoming signals in a loop.
func (c *SocketController) handleConnection(s socket.Socket, channelID, userID string) error {
	for {
		var sig request.Signal
		if err := s.Read(&sig); err != nil {
			return fmt.Errorf("failed to read signal: %w", err)
		}
		res, err := c.route(sig, channelID, userID)
		if err != nil {
			return fmt.Errorf("failed to route signal: %w", err)
		}
		if err := s.Write(res); err != nil {
			return fmt.Errorf("failed to write response: %w", err)
		}
	}
}

// route directs a parsed request based on its type.
func (c *SocketController) route(signal request.Signal, channelID, userID string) (string, error) {
	log.Println("Signal received:", signal.Type)
	switch signal.Type {
	case PUSH:
		return c.coordinator.Push(channelID, userID, signal.SDP)
	case PULL:
		return c.coordinator.Pull(channelID, userID, signal.SDP)
	case FORWARD:
		return c.coordinator.Forward(channelID, userID, signal.SDP)
	case FETCH:
		return c.coordinator.Fetch(channelID, userID, signal.SDP)
	case ARRANGE:
		return c.coordinator.Arrange(channelID, userID, signal.SDP)
	case RECONNECT:
		return c.coordinator.Reconnect(channelID, userID, signal.SDP)
	default:
		return "", fmt.Errorf("unknown request type: %s", signal.Type)
	}
}
