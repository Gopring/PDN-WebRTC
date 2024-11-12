// Package socket provides an interface for managing socket.
package socket

import (
	"github.com/gorilla/websocket"
	"net/http"
)

type WebSocket struct {
	conn *websocket.Conn
}

func New(w http.ResponseWriter, r *http.Request) (*WebSocket, error) {
	ug := websocket.Upgrader{}
	conn, err := ug.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	return &WebSocket{
		conn: conn,
	}, nil
}

func (s *WebSocket) Close() error {
	return s.conn.Close()
}

func (s *WebSocket) Write(data string) error {
	if err := s.conn.WriteMessage(websocket.TextMessage, []byte(data)); err != nil {
		return err
	}
	return nil
}

func (s *WebSocket) Read(v any) error {
	return s.conn.ReadJSON(v)
}
