// Package coordinator contains handling socket more clearly
package coordinator

import (
	"pdn/pkg/socket"
)

// Coordinator is an interface for managing socket.
//
//go:generate mockgen -destination=mock_coordinator.go -package=coordinator . Coordinator
type Coordinator interface {
	Activate(channelID string, userID string, s socket.Socket) error
	Remove(channelID, userID string) error
	Push(channelID, userID, sdp string) (string, error)
	Pull(channelID, userID, sdp string) (string, error)
	Forward(channelID, userID, sdp string) (string, error)
	Fetch(channelID, userID, sdp string) (string, error)
	Arrange(channelID, userID, sdp string) (string, error)
	Reconnect(channelID, userID, sdp string) (string, error)
}
