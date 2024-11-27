// Package socket provides an interface for managing socket.
package socket

import (
	"github.com/gorilla/websocket"
	"net/http"
)

// WebSocket wraps the gorilla/websocket connection.
type WebSocket struct {
	conn *websocket.Conn
}

// New creates a new WebSocket connection by upgrading the HTTP request.
func New(w http.ResponseWriter, r *http.Request) (*WebSocket, error) {
	ug := websocket.Upgrader{
		CheckOrigin: func(_ *http.Request) bool {
			return true
		},
	}

	conn, err := ug.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	return &WebSocket{
		conn: conn,
	}, nil
}

// Close closes the WebSocket connection.
func (s *WebSocket) Close() error {
	return s.conn.Close()
}

// Write sends a text message to the WebSocket connection.
func (s *WebSocket) Write(data string) error {
	if err := s.conn.WriteMessage(websocket.TextMessage, []byte(data)); err != nil {
		return err
	}
	return nil
}

// WriteJSON sends a text message to the WebSocket connection.
func (s *WebSocket) WriteJSON(data any) error {
	if err := s.conn.WriteJSON(data); err != nil {
		return err
	}
	return nil
}

// Read reads a JSON message from the WebSocket connection and unmarshals it into the provided variable.
func (s *WebSocket) Read(v any) error {
	return s.conn.ReadJSON(v)
}
