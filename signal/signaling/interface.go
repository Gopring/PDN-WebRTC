package signaling

import "pdn/types/api/request"

// Signal is an interface for signaling.
//
//go:generate mockgen -destination=mock_signal.go -package=signaling . Signal
type Signal interface {
	Send(signal request.Signal) (string, error)
	Receive(signal request.Signal) (string, error)
	Forward(signal request.Signal) (string, error)
	Fetch(signal request.Signal) (string, error)
	Arrange(signal request.Signal) (string, error)
	Reconnect(signal request.Signal) (string, error)
}
