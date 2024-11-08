package signaling

import (
	"pdn/media"
	"pdn/signal/coordinator"
	"pdn/types/api/request"
)

type Signaler struct {
	media       *media.Media
	coordinator *coordinator.Coordinator
}

func New(m *media.Media) *Signaler {
	return &Signaler{
		media: m,
	}
}

func (s *Signaler) Send(signal request.Signal) (string, error) {
	return s.media.AddSender(signal.ChannelID, signal.UserID, signal.SDP)
}

func (s *Signaler) Receive(signal request.Signal) (string, error) {
	return s.media.AddReceiver(signal.ChannelID, signal.UserID, signal.SDP)
}

func (s *Signaler) Forward(signal request.Signal) (string, error) {
	return s.media.AddForwarder(signal.ChannelID, signal.UserID, signal.SDP)
}

func (s *Signaler) Fetch(signal request.Signal) (string, error) {
	forwarderID, err := s.media.GetForwarder(signal.ChannelID)
	if err != nil {
		return "", err
	}
	if sdp, err := s.coordinator.RequestResponse(signal.ChannelID, forwarderID, "arrange"); err != nil {
		return "", err
	} else {
		return sdp, nil
	}
}

func (s *Signaler) Arrange(signal request.Signal) (string, error) {
	err := s.coordinator.Deliver(signal.ChannelID, signal.UserID, signal.SDP)
	if err != nil {
		return "", err
	}
	return "", nil
}

func (s *Signaler) Reconnect(signal request.Signal) (string, error) {
	return s.media.AddReceiver(signal.ChannelID, signal.UserID, signal.SDP)
}
