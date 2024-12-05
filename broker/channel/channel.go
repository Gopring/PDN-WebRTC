// Package channel provides the implementation of message channels.
package channel

import (
	"pdn/broker/subscription"
	"sync"
)

// Channel represents a message channel that can have multiple subscribers.
type Channel struct {
	mu   sync.RWMutex
	subs []*subscription.Subscription
}

// New creates and initializes a new Channel instance.
func New() *Channel {
	return &Channel{
		subs: make([]*subscription.Subscription, 0),
	}
}

// SendAll sends a message to all Channel.
func (c *Channel) SendAll(message any) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, sub := range c.subs {
		// Send message in a non-blocking manner
		go sub.Send(message)
	}
}

// AddSubscription adds a new Subscription Channel.
func (c *Channel) AddSubscription(sub *subscription.Subscription) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.subs = append(c.subs, sub)
}

// RemoveSubscription removes a Subscription Channel.
func (c *Channel) RemoveSubscription(sub *subscription.Subscription) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, s := range c.subs {
		if s == sub {
			c.subs = append(c.subs[:i], c.subs[i+1:]...)
			sub.Close()
			return
		}
	}
}
