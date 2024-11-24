// Package controller handles HTTP logic.
package controller

import (
	"fmt"
	"log"
	"net/http"
	"pdn/pkg/socket"
	"pdn/signal/coordinator"
	"pdn/types/api/request"
)

const (
	send      = "send"
	receive   = "receive"
	forward   = "forward"
	fetch     = "fetch"
	arrange   = "arrange"
	reconnect = "reconnect"
)

// SocketController handles HTTP requests.
type SocketController struct {
	coordinator coordinator.Coordinator
	debug       bool
}

// New creates a new instance of Handler.
func New(c *coordinator.MemoryCoordinator, isDebug bool) *SocketController {
	return &SocketController{
		coordinator: c,
		debug:       isDebug,
	}
}

// ServeHTTP handles HTTP requests.
func (c *SocketController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s, err := socket.New(w, r)
	if err != nil {
		log.Println(err)
		return
	}
	req := request.Activate{}
	if err := s.Read(&req); err != nil {
		log.Println(err)
		return
	}

	if err := c.coordinator.Activate(req.ChannelID, req.UserID, s); err != nil {
		log.Println(err)
		return
	}

	defer func(sk *socket.WebSocket) {
		if err := c.coordinator.Remove(req.ChannelID, req.UserID); err != nil {
			log.Println(err)
			return
		}
		if err := sk.Close(); err != nil {
			log.Println(err)
			return
		}
	}(s)

	channelID := req.ChannelID
	userID := req.UserID
	for {
		sig := request.Signal{}
		if err := s.Read(&sig); err != nil {
			log.Println(err)
			return
		}
		res, err := c.route(sig, channelID, userID)
		if err != nil {
			log.Println(err)
			return
		}
		if err = s.Write(res); err != nil {
			log.Println(err)
			return
		}
	}
}

// route routes a parsed request based on its type
func (c *SocketController) route(signal request.Signal, channelID, userID string) (string, error) {
	switch signal.Type {
	case send:
		return c.coordinator.Send(channelID, userID, signal.SDP)
	case receive:
		return c.coordinator.Receive(channelID, userID, signal.SDP)
	case forward:
		return c.coordinator.Forward(channelID, userID, signal.SDP)
	case fetch:
		return c.coordinator.Fetch(channelID, userID, signal.SDP)
	case arrange:
		return c.coordinator.Arrange(channelID, userID, signal.SDP)
	case reconnect:
		return c.coordinator.Reconnect(channelID, userID, signal.SDP)
	default:
		return "", fmt.Errorf("unknown request type: %s", signal.Type)
	}
}
