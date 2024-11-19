// Package controller handles HTTP logic.
package controller

import (
	"fmt"
	"log"
	"net/http"
	"pdn/signal/controller/socket"
	"pdn/signal/coordinator"
	"pdn/signal/signaling"
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
	signaler    signaling.Signal
	debug       bool
}

// New creates a new instance of Handler.
func New(s signaling.Signal, c *coordinator.MemoryCoordinator, isDebug bool) *SocketController {
	return &SocketController{
		signaler:    s,
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

	if err := c.coordinator.AddUser(req.ChannelID, req.UserID, s); err != nil {
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

	for {

		sig := request.Signal{}
		if err := s.Read(&sig); err != nil {
			log.Println(err)
			return
		}
		res, err := c.route(sig)
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
func (c *SocketController) route(signal request.Signal) (string, error) {
	switch signal.Type {
	case send:
		return c.signaler.Send(signal)
	case receive:
		return c.signaler.Receive(signal)
	case forward:
		return c.signaler.Forward(signal)
	case fetch:
		return c.signaler.Fetch(signal)
	case arrange:
		return c.signaler.Arrange(signal)
	case reconnect:
		return c.signaler.Reconnect(signal)
	default:
		return "", fmt.Errorf("unknown request type: %s", signal.Type)
	}
}
