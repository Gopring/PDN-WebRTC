// Package controller handles HTTP logic.
package controller

import (
	"fmt"
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

// Controller handles HTTP requests.
type Controller struct {
	coordinator *coordinator.Coordinator
	signaler    *signaling.Signaler
	debug       bool
}

// New creates a new instance of Handler.
func New(s *signaling.Signaler, isDebug bool) *Controller {
	return &Controller{
		signaler: s,
		debug:    isDebug,
	}
}

// ServeHTTP handles HTTP requests.
func (c *Controller) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s, err := socket.New(w, r)
	if err != nil {
		return
	}
	sig := request.Signal{}
	if err := s.Read(&sig); err != nil {

	}

	if err := c.coordinator.AddUser(sig.ChannelID, sig.UserID, s); err != nil {

	}

	defer func(sk *socket.Socket) {
		if err := c.coordinator.Remove(sig.ChannelID, sig.UserID); err != nil {

		}
		if err := sk.Close(); err != nil {

		}
	}(s)

	for {
		sig := request.Signal{}
		if err := s.Read(&sig); err != nil {

		}
		res, err := c.route(sig)
		if err != nil {
			return
		}
		if err = s.Write(res); err != nil {

		}
	}
}

// route routes a parsed request based on its type
func (c *Controller) route(signal request.Signal) (string, error) {
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
