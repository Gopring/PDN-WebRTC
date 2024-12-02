// Package media contains managing channels and connections using WebRTC.
package media

import (
	"fmt"
	"github.com/pion/webrtc/v4"
	"log"
	"pdn/broker"
	"pdn/types/client/response"
	"pdn/types/message"
)

// Media contains the channels and connection configuration.
// NOTE(window9u): In future, media package could be detached from pdn
// and be used as a standalone package.
type Media struct {
	// TODO(window9u): we should add locker for channels.
	broker           *broker.Broker
	channels         map[string]*Channel
	connectionConfig webrtc.Configuration
}

// New creates a new Media instance.
// TODO(window9u): we should add more configuration options.
func New(b *broker.Broker) *Media {
	return &Media{
		broker:   b,
		channels: map[string]*Channel{},
		connectionConfig: webrtc.Configuration{
			ICEServers: []webrtc.ICEServer{
				{
					URLs: []string{"stun:stun.l.google.com:19302"},
				},
			},
		},
	}
}

func (m *Media) Run() {
	pushEvent := m.broker.Subscribe(broker.ClientMessage, broker.PUSH)
	pullEvent := m.broker.Subscribe(broker.ClientMessage, broker.PULL)

	for {
		select {
		case event := <-pushEvent:
			push, ok := event.(message.Push)
			if !ok {
				log.Println("Failed to cast event to push")
				break
			}
			serverSDP, err := m.AddUpstream(push.ChannelID, push.UserID, push.SDP)
			if err != nil {
				log.Printf("Failed to add upstream: %v", err)
				break
			}
			if err := m.broker.Publish(broker.ClientSocket, broker.Detail(push.ChannelID+push.UserID), response.Push{
				RequestID:  push.RequestID,
				StatusCode: 200,
				SDP:        serverSDP,
			}); err != nil {
				log.Printf("Failed to publish to broker: %v", err)
			}
		case event := <-pullEvent:
			pull, ok := event.(message.Pull)
			if !ok {
				log.Println("Failed to cast event to pull")
				break
			}
			serverSDP, err := m.AddDownstream(pull.ChannelID, pull.UserID, pull.SDP)
			if err != nil {
				log.Printf("Failed to add downstream: %v", err)
				break
			}
			if err := m.broker.Publish(broker.ClientSocket, broker.Detail(pull.ChannelID+pull.UserID), response.Pull{
				RequestID:  pull.RequestID,
				StatusCode: 200,
				SDP:        serverSDP,
			}); err != nil {
				log.Printf("Failed to publish to broker: %v", err)
			}
		}
	}
}

// AddUpstream creates a new upstream connection and adds it to the channel.
func (m *Media) AddUpstream(channelID string, userID string, sdp string) (string, error) {
	ch := NewChannel()
	conn, err := NewInbound(m.connectionConfig)
	if err != nil {
		return "", fmt.Errorf("failed to make connection: %w", err)
	}
	conn.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		log.Printf("ICE Connection State has changed: %s\n", state.String())
		if state == webrtc.PeerConnectionStateConnected {
			if err := m.broker.Publish(broker.Media, broker.PUSH, message.Connected{
				ChannelID: channelID,
				UserID:    userID,
			}); err != nil {
				log.Printf("Failed to publish to broker: %v", err)
			}
		} else if state == webrtc.PeerConnectionStateClosed {
			if err := m.broker.Publish(broker.Media, broker.PUSH, message.Disconnected{
				ChannelID: channelID,
				UserID:    userID,
			}); err != nil {
				log.Printf("Failed to publish to broker: %v", err)
			}
		}
	})

	ch.SetUpstream(conn, userID)
	if err = StartICE(conn, sdp); err != nil {
		return "", fmt.Errorf("failed to start ICE: %w", err)
	}
	m.channels[channelID] = ch

	return conn.LocalDescription().SDP, nil
}

// AddDownstream creates a new downstream connection and adds it to the channel.
func (m *Media) AddDownstream(channelID string, userID string, sdp string) (string, error) {
	ch := m.channels[channelID]
	conn, err := NewOutbound(m.connectionConfig)
	if err != nil {
		return "", fmt.Errorf("failed to make connection: %w", err)
	}

	conn.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		log.Printf("ICE Connection State has changed: %s\n", state.String())
		if state == webrtc.PeerConnectionStateConnected {
			if err := m.broker.Publish(broker.Media, broker.PUSH, message.Connected{
				ChannelID: channelID,
				UserID:    userID,
			}); err != nil {
				log.Printf("Failed to publish to broker: %v", err)
			}
		} else if state == webrtc.PeerConnectionStateClosed {
			if err := m.broker.Publish(broker.Media, broker.PUSH, message.Disconnected{
				ChannelID: channelID,
				UserID:    userID,
			}); err != nil {
				log.Printf("Failed to publish to broker: %v", err)
			}
		}
	})

	if err = ch.SetDownstream(conn, userID); err != nil {
		return "", fmt.Errorf("failed to set downstream: %w", err)
	}

	if err = StartICE(conn, sdp); err != nil {
		return "", fmt.Errorf("failed to start ICE: %w", err)
	}
	if err != nil {
		return "", fmt.Errorf("failed to start ICE: %w", err)
	}
	return conn.LocalDescription().SDP, nil
}
