// Package socket provides an interface for managing socket.
package handler

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"pdn/signal/controller"
)

// Handler wraps the gorilla/websocket connection.
type Handler struct {
	controller *controller.Controller
}

// New creates a new SocketHandler connection by upgrading the HTTP request.
func New(c *controller.Controller) *Handler {
	return &Handler{
		controller: c,
	}
}

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
	defer func(conn *websocket.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println("Failed to close connection")
			return
		}
	}(conn)
	h.controller.Process(conn)
}
