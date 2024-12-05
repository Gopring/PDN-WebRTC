// Package subscription provides a mechanism for managing message subscriptions.
package subscription

// Subscription represents a message subscription.
type Subscription struct {
	queue chan any
}

// New creates and initializes a new Subscription instance.
func New() *Subscription {
	return &Subscription{
		queue: make(chan any, 1),
	}
}

// Send enqueues a message to the subscription.
func (s *Subscription) Send(message any) {
	s.queue <- message
}

// Receive retrieves the channel for incoming messages.
func (s *Subscription) Receive() <-chan any {
	return s.queue
}

// Close closes the subscription, releasing any resources.
func (s *Subscription) Close() {
	close(s.queue)
}
