// Package media manages channels and connections using WebRTC.
package media

import (
	"fmt"
	"log"
	"sync"

	"github.com/pion/webrtc/v4"
	"pdn/broker"
	"pdn/types/client/response"
	"pdn/types/message"
)

// Media manages channels and connection configurations.
// NOTE: In the future, the media package could be detached from pdn
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
// TODO: Add more configuration options.
func New(b *broker.Broker) *Media {
	return &Media{
		broker:           b,
		channels:         make(map[string]*Channel),
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
		return fmt.Errorf("failed to cast event to Push: %v", event)
	}
	serverSDP, err := m.AddUpstream(push.ChannelID, push.UserID, push.SDP)
	if err != nil {
		return fmt.Errorf("failed to add upstream: %w", err)
	}
	if err := m.broker.Publish(broker.ClientSocket, broker.Detail(push.ChannelID+push.UserID), response.Push{
		RequestID:  push.RequestID,
		StatusCode: 200,
		SDP:        serverSDP,
	}); err != nil {
		return fmt.Errorf("failed to publish push response: %w", err)
	}
	return nil
}

func (m *Media) handlePull(event any) error {
	pull, ok := event.(message.Pull)
	if !ok {
		return fmt.Errorf("failed to cast event to Pull: %v", event)
	}
	serverSDP, err := m.AddDownstream(pull.ChannelID, pull.SDP)
	if err != nil {
		return fmt.Errorf("failed to add downstream: %w", err)
	}
	if err := m.broker.Publish(broker.ClientSocket, broker.Detail(pull.ChannelID+pull.UserID), response.Pull{
		RequestID:  pull.RequestID,
		StatusCode: 200,
		SDP:        serverSDP,
	}); err != nil {
		return fmt.Errorf("failed to publish pull response: %w", err)
	}
	return nil
}

// AddUpstream creates a new upstream connection and adds it to the channel.
func (m *Media) AddUpstream(channelID, userID, sdp string) (string, error) {
	if m.channelExists(channelID) {
		return "", fmt.Errorf("channel already exists: %s", channelID)
	}

	ch := NewChannel()
	m.addChannel(channelID, ch)

	conn, err := NewInboundConnection(m.connectionConfig)
	if err != nil {
		return "", fmt.Errorf("failed to create inbound connection: %w", err)
	}
	m.publishStateChange(conn, channelID)

	ch.SetUpstream(conn, userID)
	if err = StartICE(conn, sdp); err != nil {
		return "", fmt.Errorf("failed to start ICE: %w", err)
	}

	return conn.LocalDescription().SDP, nil
}

// AddDownstream creates a new downstream connection and adds it to the channel.
func (m *Media) AddDownstream(channelID, sdp string) (string, error) {
	ch, err := m.getChannel(channelID)
	if err != nil {
		return "", err
	}

	conn, err := NewOutboundConnection(m.connectionConfig)
	if err != nil {
		return "", fmt.Errorf("failed to create outbound connection: %w", err)
	}

	m.publishStateChange(conn, channelID)

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

func (m *Media) getChannel(channelID string) (*Channel, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	ch, ok := m.channels[channelID]
	if !ok {
		return nil, fmt.Errorf("channel does not exist: %s", channelID)
	}
	return ch, nil
}

func (m *Media) publishStateChange(conn *webrtc.PeerConnection, channelID string) {
	conn.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		log.Printf("Channel %s: ICE Connection State has changed to %s", channelID, state.String())
		switch state {
		case webrtc.PeerConnectionStateConnected:
			// TODO: Publish this event to the broker.
			log.Printf("Channel %s: Connected", channelID)
		case webrtc.PeerConnectionStateClosed:
			// TODO: Publish this event to the broker.
			log.Printf("Channel %s: Closed", channelID)
		case webrtc.PeerConnectionStateDisconnected:
			log.Printf("Channel %s: Disconnected", channelID)
		default:
			panic("unhandled default case")
		}
	})
}
