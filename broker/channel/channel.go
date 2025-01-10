// Package channel manages message channels.
package channel

import (
	"log"
	"pdn/broker/subscription"
	"sync"
	"time"
)

// Channel represents a message channel.
type Channel struct {
	mu     sync.RWMutex
	topic  string
	detail string
	subs   []*subscription.Subscription
}

// New creates a new Channel instance.
func New(topic, detail string) *Channel {
	return &Channel{
		topic:  topic,
		detail: detail,
		subs:   make([]*subscription.Subscription, 0),
	}
}

// SendAll sends a message to all Channel.
func (c *Channel) SendAll(message any) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, sub := range c.subs {
		select {
		case sub.Send() <- message:
		case <-time.After(1 * time.Second):
			log.Printf("Timeout occurs in sending message to Topic: %s, Detail: %s, Message:%v", c.topic, c.detail, message)
		}
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
