// Package coordinator contains handling socket more clearly
package coordinator

import (
	"fmt"
	"pdn/pkg/socket"
	"time"
)

// User controls socket and relay message
type User struct {
	//TODO(window9u): we should add lock for concurrent access
	socket   socket.Socket
	response chan string
}

// Request send data to user
func (u *User) Request(data any) error {
	err := u.socket.WriteJSON(data)
	if err != nil {
		return err
	}
	return nil
}

// WaitForResponse wait response for duration. and returns response
func (u *User) WaitForResponse(d time.Duration) (string, error) {
	select {
	case sdp := <-u.response:
		return sdp, nil
	case <-time.After(d):
		return "", fmt.Errorf("timeout")
	}
}

// Response send data to user
func (u *User) Response(data string, d time.Duration) error {
	select {
	case <-time.After(d):
		return fmt.Errorf("timeout")
	case u.response <- data:
		return nil
	}
}
