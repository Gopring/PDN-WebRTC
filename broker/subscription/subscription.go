package subscription

type Subscription struct {
	queue chan any
}

func New() *Subscription {
	return &Subscription{
		queue: make(chan any, 1),
	}
}

func (s *Subscription) Send(message any) {
	s.queue <- message
}

func (s *Subscription) Receive() any {
	msg, ok := <-s.queue
	if !ok {
		return nil
	}
	return msg
}

func (s *Subscription) Close() {
	close(s.queue)
}
