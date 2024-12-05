// Package subscription manages message subscriptions.
package subscription

// Subscription is a message subscription.
type Subscription struct {
	queue chan any
}

// New creates a new subscription instance.
func New() *Subscription {
	return &Subscription{
		queue: make(chan any, 1),
	}
}

// Send sends a message to the subscription.
func (s *Subscription) Send() chan<- any {
	return s.queue
}

// Receive returns a channel to receive messages.
func (s *Subscription) Receive() <-chan any {
	return s.queue
}

// Close closes the subscription.
func (s *Subscription) Close() {
	close(s.queue)
}
