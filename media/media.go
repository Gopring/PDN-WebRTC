// Package media contains managing channels and connections using WebRTC.
package media

import (
	"fmt"
	"github.com/pion/webrtc/v4"
	"log"
	"pdn/broker"
	"pdn/types/client/response"
	"pdn/types/message"
	"sync"
)

// Media contains the channels and connection configuration.
// NOTE(window9u): In future, media package could be detached from pdn
// and be used as a standalone package.
type Media struct {
	mu               sync.RWMutex
	broker           *broker.Broker
	channels         map[string]*Channel
	connectionConfig webrtc.Configuration
}

var defaultWebrtcConfig = webrtc.Configuration{
	ICEServers: []webrtc.ICEServer{
		{
			URLs: []string{"stun:stun.l.google.com:19302"},
		},
	},
}

// New creates a new Media instance.
// TODO(window9u): we should add more configuration options.
func New(b *broker.Broker) *Media {
	return &Media{
		broker:           b,
		channels:         map[string]*Channel{},
		connectionConfig: defaultWebrtcConfig,
	}
}

func (m *Media) Run() {
	pushEvent := m.broker.Subscribe(broker.ClientMessage, broker.PUSH)
	pullEvent := m.broker.Subscribe(broker.ClientMessage, broker.PULL)

	for {
		var err error
		select {
		case event := <-pushEvent.Receive():
			err = m.handlePush(event)
		case event := <-pullEvent.Receive():
			err = m.handlePull(event)
		}
		if err != nil {
			log.Printf("Failed to handle event in Media: %v", err)
		}
	}
}

func (m *Media) handlePush(event any) error {
	push, ok := event.(message.Push)
	if !ok {
		return fmt.Errorf("failed to cast event to push %v", event)
	}
	serverSDP, err := m.AddUpstream(push.ChannelID, push.UserID, push.SDP)
	if err != nil {
		return fmt.Errorf("failed to add upstream: %v", err)
	}
	if err := m.broker.Publish(broker.ClientSocket, broker.Detail(push.ChannelID+push.UserID), response.Push{
		RequestID:  push.RequestID,
		StatusCode: 200,
		SDP:        serverSDP,
	}); err != nil {
		return fmt.Errorf("failed to publish push response: %v", err)
	}
	return nil
}

func (m *Media) handlePull(event any) error {
	pull, ok := event.(message.Pull)
	if !ok {
		return fmt.Errorf("failed to cast event to pull %v", event)
	}
	serverSDP, err := m.AddDownstream(pull.ChannelID, pull.SDP)
	if err != nil {
		return fmt.Errorf("failed to add downstream: %v", err)
	}
	if err := m.broker.Publish(broker.ClientSocket, broker.Detail(pull.ChannelID+pull.UserID), response.Pull{
		RequestID:  pull.RequestID,
		StatusCode: 200,
		SDP:        serverSDP,
	}); err != nil {
		return fmt.Errorf("failed to publish pull response: %v", err)
	}
	return nil
}

// AddUpstream creates a new upstream connection and adds it to the channel.
func (m *Media) AddUpstream(channelID string, userID string, sdp string) (string, error) {
	if m.channelExists(channelID) {
		return "", fmt.Errorf("channel already exists: %s", channelID)
	}

	ch := NewChannel()
	conn, err := NewInboundConnection(m.connectionConfig)
	if err != nil {
		return "", fmt.Errorf("failed to make connection: %w", err)
	}
	m.PublishStateChange(conn, channelID)

	ch.SetUpstream(conn, userID)
	if err = StartICE(conn, sdp); err != nil {
		return "", fmt.Errorf("failed to start ICE: %w", err)
	}

	return conn.LocalDescription().SDP, nil
}

// AddDownstream creates a new downstream connection and adds it to the channel.
func (m *Media) AddDownstream(channelID string, sdp string) (string, error) {
	ch := m.channels[channelID]
	conn, err := NewOutboundConnection(m.connectionConfig)
	if err != nil {
		return "", fmt.Errorf("failed to make connection: %w", err)
	}

	m.PublishStateChange(conn, channelID)

	if err = ch.SetDownstream(conn); err != nil {
		return "", fmt.Errorf("failed to set downstream: %w", err)
	}

	if err = StartICE(conn, sdp); err != nil {
		return "", fmt.Errorf("failed to start ICE: %w", err)
	}

	return conn.LocalDescription().SDP, nil
}

func (m *Media) addChannel(channelID string, ch *Channel) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.channels[channelID] = ch
}

func (m *Media) channelExists(channelID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.channels[channelID]
	return ok
}

func (m *Media) PublishStateChange(conn *webrtc.PeerConnection, channelID string) {
	conn.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		log.Printf("ICE Connection State has changed: %s\n", state.String())
		if state == webrtc.PeerConnectionStateConnected {
			// TODO(window9u): we should publish this event to broker.
			log.Printf("Connected: %s", channelID)
		} else if state == webrtc.PeerConnectionStateClosed {
			// TODO(window9u): we should publish this event to broker.
			log.Printf("Closed: %s", channelID)
		} else if state == webrtc.PeerConnectionStateDisconnected {
			log.Printf("Disconnected: %s", channelID)
		}
	})
}
