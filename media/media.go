// Package media contains managing channels and connections using WebRTC.
package media

import (
	"fmt"
	"github.com/pion/webrtc/v4"
	"pdn/media/channel"
	"pdn/media/connection"
)

// Func receives user's request and returns sdp
type Func func(channelID string, userID string, sdp string) (string, error)

// Media contains the channels and connection configuration.
// NOTE(window9u): In future, media package could be detached from pdn
// and be used as a standalone package.
type Media struct {
	// TODO(window9u): we should add locker for channels.
	channels         map[string]*channel.Channel
	connectionConfig webrtc.Configuration
}

// New creates a new Media instance.
// TODO(window9u): we should add more configuration options.
func New() *Media {
	return &Media{
		channels: map[string]*channel.Channel{},
		connectionConfig: webrtc.Configuration{
			ICEServers: []webrtc.ICEServer{
				{
					URLs: []string{"stun:stun.l.google.com:19302"},
				},
			},
		},
	}
}

// AddSender creates a new upstream connection and adds it to the channel.
func (m *Media) AddSender(channelID string, userID string, sdp string) (string, error) {
	ch := channel.New()
	conn, err := connection.NewInbound(m.connectionConfig, sdp)
	if err != nil {
		return "", fmt.Errorf("failed to make connection: %w", err)
	}

	ch.SetUpstream(conn, userID)

	err = conn.StartICE()
	if err != nil {
		return "", fmt.Errorf("failed to start ICE: %w", err)
	}
	m.channels[channelID] = ch
	return conn.ServerSDP(), nil
}

// AddReceiver creates a new downstream connection and adds it to the channel.
func (m *Media) AddReceiver(channelID string, userID string, sdp string) (string, error) {
	ch := m.channels[channelID]
	conn, err := connection.NewOutbound(m.connectionConfig, sdp)
	if err != nil {
		return "", fmt.Errorf("failed to make connection: %w", err)
	}

	if err = ch.SetDownstream(conn, userID); err != nil {
		return "", fmt.Errorf("failed to set downstream: %w", err)
	}

	err = conn.StartICE()
	if err != nil {
		return "", fmt.Errorf("failed to start ICE: %w", err)
	}

	return conn.ServerSDP(), nil
}
