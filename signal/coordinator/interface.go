// Package coordinator contains handling socket more clearly
package coordinator

import "pdn/signal/controller/socket"

// Coordinator is an interface for managing socket.
//
//go:generate mockgen -destination=mock_coordinator.go -package=coordinator . Coordinator
type Coordinator interface {
	RequestResponse(channelID string, userID string, data string) (string, error)
	Activate(channelID string, userID string, s socket.Socket) error
	Response(channelID, userID string, data string) error
	Remove(channelID, userID string) error
	Send(channelID, userID, sdp string) (string, error)
	Receive(channelID, userID, sdp string) (string, error)
	Forward(channelID, userID, sdp string) (string, error)
	Fetch(channelID, userID, sdp string) (string, error)
	Arrange(channelID, userID, sdp string) (string, error)
	Reconnect(channelID, userID, sdp string) (string, error)
}
