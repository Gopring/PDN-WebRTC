// Package broker manages message channels and subscriptions.
package broker

import (
	"fmt"
	"sync"

	"pdn/broker/channel"
	"pdn/broker/subscription"
)

// Topic represents a message topic.
type Topic int

// Detail represents a message detail.
type Detail string

// Topic constants for message topics.
const (
	ClientSocket Topic = iota
	Client
	Media
	Peer
)

// Detail constants for message details.
const (
	ACTIVATE     Detail = "ACTIVATE"
	DEACTIVATE   Detail = "DEACTIVATE"
	PUSH         Detail = "PUSH"
	PULL         Detail = "PULL"
	UPSTREAM     Detail = "UPSTREAM"
	DOWNSTREAM   Detail = "DOWNSTREAM"
	CONNECTED    Detail = "CONNECTED"
	DISCONNECTED Detail = "DISCONNECTED"
	FAILED       Detail = "FAILED"
	CLEAR        Detail = "CLEAR"
)

// Broker is a message broker that manages message channels and subscriptions.
type Broker struct {
	mu       sync.RWMutex
	channels map[Topic]map[Detail]*channel.Channel
}

// New creates a new broker instance.
func New() *Broker {
	return &Broker{
		channels: make(map[Topic]map[Detail]*channel.Channel),
	}
}

// Publish sends a message to all subscribers for a given topic and detail.
func (b *Broker) Publish(topic Topic, detail Detail, message any) error {
	ch, err := b.getChannel(topic, detail)
	if err != nil {
		return err
	}

	ch.SendAll(message)
	return nil
}

// Subscribe creates a subscription for a given topic and detail.
func (b *Broker) Subscribe(topic Topic, detail Detail) *subscription.Subscription {
	b.ensureChannel(topic, detail)

	sub := subscription.New()
	b.mu.RLock()
	defer b.mu.RUnlock()

	b.channels[topic][detail].AddSubscription(sub)
	return sub
}

// Unsubscribe removes a subscription for a given topic and detail.
func (b *Broker) Unsubscribe(topic Topic, detail Detail, sub *subscription.Subscription) error {
	ch, err := b.getChannel(topic, detail)
	if err != nil {
		return err
	}

	ch.RemoveSubscription(sub)
	return nil
}

// ensureChannel initializes the channel for a given topic and detail if it doesn't exist.
func (b *Broker) ensureChannel(topic Topic, detail Detail) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.channels[topic]; !exists {
		b.channels[topic] = make(map[Detail]*channel.Channel)
	}
	if _, exists := b.channels[topic][detail]; !exists {
		b.channels[topic][detail] = channel.New(topic.String(), string(detail))
	}
}

// getChannel safely retrieves the channel for a given topic and detail.
func (b *Broker) getChannel(topic Topic, detail Detail) (*channel.Channel, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if details, exists := b.channels[topic]; exists {
		if ch, exists := details[detail]; exists {
			return ch, nil
		}
	}
	return nil, fmt.Errorf("channel does not exist for topic %s and detail %s", topic.String(), detail)
}

// String returns the string representation of the Topic.
func (t Topic) String() string {
	switch t {
	case ClientSocket:
		return "ClientSocket"
	case Client:
		return "Client"
	case Media:
		return "Media"
	default:
		return "Unknown"
	}
}
