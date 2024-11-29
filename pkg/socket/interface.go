// Package socket provides an interface for managing socket.
package socket

// Socket is an interface for managing socket.
//
//go:generate mockgen -destination=mock_socket.go -package=socket . Socket
type Socket interface {
	Close() error
	WriteJSON(data any) error
	ReadJSON(v any) error
}
