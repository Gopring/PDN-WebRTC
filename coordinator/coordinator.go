// Package coordinator provides mechanisms to coordinate communication and data flow
// between clients using the broker system.
package coordinator

import (
	"pdn/broker"
)

// Coordinator is responsible for managing and coordinating communication
// between multiple clients through the broker.
type Coordinator struct {
	broker *broker.Broker
}

// Client represents an individual client in the coordination system.
//type Client struct {
//	forwardTo map[string]*Client
//	fetchFrom map[string]*Client
//}

// New creates and initializes a new Coordinator instance.
func New(b *broker.Broker) *Coordinator {
	return &Coordinator{
		broker: b,
	}
}

//func (p *Coordinator) Run() error {
//}
