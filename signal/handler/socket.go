// Package handler provides an interface for managing socket.
package handler

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"pdn/metric"
	"pdn/signal/controller"
)

// Handler wraps the gorilla/websocket connection.
type Handler struct {
	controller *controller.Controller
	metrics    *metric.Metrics
}

// New creates a new SocketHandler connection by upgrading the HTTP request.
func New(c *controller.Controller, m *metric.Metrics) *Handler {
	return &Handler{
		controller: c,
		metrics:    m,
	}
}

// ServeHTTP handles the HTTP request and upgrades it to websocket connection.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ug := websocket.Upgrader{
		CheckOrigin: func(_ *http.Request) bool {
			return true
		},
	}

	conn, err := ug.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	h.metrics.IncrementWebSocketConnections()
	defer func(conn *websocket.Conn) {
		h.metrics.DecrementWebSocketConnections()
		err := conn.Close()
		if err != nil {
			log.Println("Error occurs in closing connection")
			return
		}
	}(conn)
	if err := h.controller.Process(conn); err != nil {
		log.Printf("Error occurs in connection %v", err)
	}
}
