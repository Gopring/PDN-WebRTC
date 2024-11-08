package socket

import (
	"github.com/gorilla/websocket"
	"net/http"
)

type Socket struct {
	conn *websocket.Conn
}

func New(w http.ResponseWriter, r *http.Request) (*Socket, error) {
	ug := websocket.Upgrader{}
	conn, err := ug.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	return &Socket{
		conn: conn,
	}, nil
}

func (s *Socket) Close() error {
	return s.conn.Close()
}

func (s *Socket) Write(data string) error {
	if err := s.conn.WriteMessage(websocket.TextMessage, []byte(data)); err != nil {
		return err
	}
	return nil
}

func (s *Socket) Read(v any) error {
	return s.conn.ReadJSON(v)
}
