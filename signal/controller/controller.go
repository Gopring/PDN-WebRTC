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

// Controller handles HTTP requests.
type Controller struct {
	coordinator *coordinator.Coordinator
	signaler    *signaling.Signaler
	debug       bool
}

// New creates a new instance of Handler.
func New(s *signaling.Signaler, c *coordinator.Coordinator, isDebug bool) *Controller {
	return &Controller{
		signaler:    s,
		coordinator: c,
		debug:       isDebug,
	}
}

// ServeHTTP handles HTTP requests.
func (c *Controller) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s, err := socket.New(w, r)
	if err != nil {
		log.Println(err)
		return
	}
	sig := request.Signal{}
	if err := s.Read(&sig); err != nil {
		log.Println(err)
		return
	}

	if err := c.coordinator.AddUser(sig.ChannelID, sig.UserID, s); err != nil {
		log.Println(err)
		return
	}

	defer func(sk *socket.Socket) {
		if err := c.coordinator.Remove(sig.ChannelID, sig.UserID); err != nil {
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
