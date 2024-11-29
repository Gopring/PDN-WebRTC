// Package media contains managing channels and connections using WebRTC.
package media

import (
	"fmt"
	"github.com/pion/webrtc/v4"
	"log"
	"pdn/broker"
	"pdn/media/channel"
	"pdn/media/connection"
	"pdn/types/message"
)

// Func receives user's request and returns sdp
type Func func(channelID string, userID string, sdp string) (string, error)

// Media contains the channels and connection configuration.
// NOTE(window9u): In future, media package could be detached from pdn
// and be used as a standalone package.
type Media struct {
	// TODO(window9u): we should add locker for channels.
	broker           broker.Broker
	channels         map[string]*channel.Channel
	connectionConfig webrtc.Configuration
}

// New creates a new Media instance.
// TODO(window9u): we should add more configuration options.
func New(b broker.Broker) *Media {
	return &Media{
		broker:   b,
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

func (m *Media) Run() error {
	pushEvent, err := m.broker.Subscribe(broker.PUSH, broker.COORDINATOR)
	if err != nil {
		return err
	}
	pullEvent, err := m.broker.Subscribe(broker.PULL, broker.COORDINATOR)
	if err != nil {
		return err
	}
	for {
		select {
		case event := <-pushEvent:
			push, ok := event.(message.Push)
			if !ok {
				log.Println("Failed to cast event to push")
				continue
			}
			remoteSDP, err := m.AddSender(push.ChannelID, push.UserID, push.SDP)
			if err != nil {
				log.Printf("Failed to add sender: %v", err)
				continue
			}
			if err := m.broker.Send(broker.PUSH, push.UserID+push.ChannelID, remoteSDP); err != nil {
				log.Printf("Failed to publish to broker: %v", err)
			}
		case event := <-pullEvent:
			pull, ok := event.(message.Pull)
			if !ok {
				log.Println("Failed to cast event to pull")
				continue
			}
			remoteSDP, err := m.AddReceiver(pull.ChannelID, pull.UserID, pull.SDP)
			if err != nil {
				log.Printf("Failed to add receiver: %v", err)
				continue
			}
			if err := m.broker.Send(broker.PULL, pull.UserID+pull.ChannelID, remoteSDP); err != nil {
				log.Printf("Failed to publish to broker: %v", err)
			}
		}
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

	if err = conn.StartICE(); err != nil {
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
