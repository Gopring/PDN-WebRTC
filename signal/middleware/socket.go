// Package middleware contains common middleware functions for HTTP handlers.
package middleware

import (
	"bufio"
	"errors"
	"log"
	"net"
	"net/http"
	"pdn/pkg/socket"
	"pdn/signal/controller"
)

// Socket creates socket and call controller.
type Socket struct {
	http.ResponseWriter
	controller controller.Controller
}

// NewSocket creates a new Socket middleware.
func NewSocket(con controller.Controller) *Socket {
	return &Socket{
		controller: con,
	}
}

// Hijack hijacks the connection. This is necessary for using websockets.
func (s *Socket) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := s.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("hijack not supported")
	}
	return h.Hijack()
}

// Intercept processes the request and call the next handler.
func (s *Socket) Intercept(next http.Handler) http.Handler {
	con := s.controller
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s, err := socket.New(w, r)
		if err != nil {
			log.Printf("Failed to create WebSocket: %v", err)
			return
		}
		if err := con.Process(s); err != nil {
			log.Printf("Failed to process WebSocket: %v", err)
			return
		}
		defer func() {
			if err := s.Close(); err != nil {
				log.Printf("Failed to close WebSocket: %v", err)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
