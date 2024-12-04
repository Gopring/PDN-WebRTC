package coordinator

import (
	"pdn/broker"
)

type Coordinator struct {
	broker *broker.Broker
}

type Client struct {
	forwardTo map[string]*Client
	fetchFrom map[string]*Client
}

func New(b *broker.Broker) *Coordinator {
	return &Coordinator{
		broker: b,
	}
}

//func (p *Coordinator) Run() error {
//}
