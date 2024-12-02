package broker

import (
	"errors"
	"sync"
)

const (
	ClientSocket = iota
	ClientMessage
	Media
)

const (
	PUSH = "PUSH"
	PULL = "PULL"
)

type Topic int
type Detail string

type Broker struct {
	mu     sync.RWMutex
	queues map[Topic]map[Detail]*subscriptions
}

type subscriptions struct {
	mu            sync.RWMutex
	subscriptions []chan any
}

// New creates a new broker instance.
func New() *Broker {
	return &Broker{
		queues: make(map[Topic]map[Detail]*subscriptions),
	}
}

// Publish publishes a message to all subscribers for a given topic and detail.
func (b *Broker) Publish(topic Topic, detail Detail, message any) error {
	subs, err := b.getSubscriptions(topic, detail)
	if err != nil {
		return err
	}

	subs.sendAll(message)
	return nil
}

// Subscribe creates a subscription for a given topic and detail and returns a channel.
func (b *Broker) Subscribe(topic Topic, detail Detail) <-chan any {
	if _, exists := b.queues[topic]; !exists {
		b.queues[topic] = make(map[Detail]*subscriptions)
	}
	if _, exists := b.queues[topic][detail]; !exists {
		b.queues[topic][detail] = &subscriptions{}
	}

	subscription := make(chan any, 1)
	b.queues[topic][detail].addSubscription(subscription)
	return subscription
}

// Unsubscribe removes a subscription for a given topic and detail.
func (b *Broker) Unsubscribe(topic Topic, detail Detail, ch chan any) error {
	subs, err := b.getSubscriptions(topic, detail)
	if err != nil {
		return err
	}

	subs.removeSubscription(ch)
	return nil
}

// getSubscriptions safely retrieves subscriptions for a given topic and detail.
func (b *Broker) getSubscriptions(topic Topic, detail Detail) (*subscriptions, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if details, exists := b.queues[topic]; exists {
		if subs, exists := details[detail]; exists {
			return subs, nil
		}
	}
	return nil, errors.New("subscriptions do not exist")
}

// subscriptionExists checks if a subscription exists for the given topic and detail.
func (b *Broker) subscriptionExists(topic Topic, detail Detail) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	_, exists := b.queues[topic][detail]
	return exists
}

// sendAll sends a message to all subscriptions.
func (s *subscriptions) sendAll(message any) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, ch := range s.subscriptions {
		// Send message in a non-blocking manner
		go func(ch chan any) {
			ch <- message
		}(ch)
	}
}

// addSubscription adds a new subscription channel.
func (s *subscriptions) addSubscription(ch chan any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.subscriptions = append(s.subscriptions, ch)
}

// removeSubscription removes a subscription channel.
func (s *subscriptions) removeSubscription(ch chan any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, sub := range s.subscriptions {
		if sub == ch {
			s.subscriptions = append(s.subscriptions[:i], s.subscriptions[i+1:]...)
			close(ch)
			return
		}
	}
}
