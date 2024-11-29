package classifier

import "pdn/broker"

type Classifier struct {
	broker *broker.Broker
}

func New(b *broker.Broker) *Classifier {
	return &Classifier{
		broker: b,
	}
}
