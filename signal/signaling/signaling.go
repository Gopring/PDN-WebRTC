// Package signaling contains logic for handling signaling processes between media and coordinator modules.
package signaling

import (
	"pdn/media"
	"pdn/signal/coordinator"
	"pdn/types/api/request"
)

// SignalHandler is a struct for signaling.
type SignalHandler struct {
	media       media.Media
	coordinator coordinator.Coordinator
}

// New creates a new instance of SignalHandler.
func New(m media.Media, cod coordinator.Coordinator) *SignalHandler {
	return &SignalHandler{
		media:       m,
		coordinator: cod,
	}
}

// Send sends a signal.
func (s *SignalHandler) Send(signal request.Signal) (string, error) {
	return s.media.AddSender(signal.ChannelID, signal.UserID, signal.SDP)
}

// Receive receives a signal.
func (s *SignalHandler) Receive(signal request.Signal) (string, error) {
	return s.media.AddReceiver(signal.ChannelID, signal.UserID, signal.SDP)
}

// Forward forwards a signal.
func (s *SignalHandler) Forward(signal request.Signal) (string, error) {
	return s.media.AddForwarder(signal.ChannelID, signal.UserID, signal.SDP)
}

// Fetch fetches a signal.
func (s *SignalHandler) Fetch(signal request.Signal) (string, error) {
	forwarderID, err := s.media.GetForwarder(signal.ChannelID)
	if err != nil {
		return "", err
	}
	sdp, err := s.coordinator.RequestResponse(signal.ChannelID, forwarderID, "arrange")
	if err != nil {
		return "", err
	}
	return sdp, nil
}

// Arrange arranges a signal.
func (s *SignalHandler) Arrange(signal request.Signal) (string, error) {
	err := s.coordinator.Response(signal.ChannelID, signal.UserID, signal.SDP)
	if err != nil {
		return "", err
	}
	return "", nil
}

// Reconnect reconnects a signal.
func (s *SignalHandler) Reconnect(signal request.Signal) (string, error) {
	return s.media.AddReceiver(signal.ChannelID, signal.UserID, signal.SDP)
}
