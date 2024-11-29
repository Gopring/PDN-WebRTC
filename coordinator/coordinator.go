package coordinator

import (
	"log"
	"pdn/broker"
	"pdn/types/message"
)

type Coordinator struct {
	broker *broker.Broker
}

func New(b *broker.Broker) *Coordinator {
	return &Coordinator{
		broker: b,
	}
}

func (p *Coordinator) Run() error {
	pushEvent, err := p.broker.Subscribe(broker.PUSH, broker.CONTROLLER)
	if err != nil {
		return err
	}
	pullEvent, err := p.broker.Subscribe(broker.PULL, broker.CONTROLLER)
	if err != nil {
		return err
	}
	for {
		select {
		case event := <-pushEvent:
			push, ok := event.(message.Push)
			if !ok {
				log.Println("Failed to cast event to push")
				continue
			}
			if err := p.broker.Publish(broker.PUSH, broker.COORDINATOR, push.SDP); err != nil {
				log.Printf("Failed to publish to broker: %v", err)
			}
		case event := <-pullEvent:
			pull, ok := event.(message.Pull)
			if !ok {
				log.Println("Failed to cast event to pull")
				continue
			}
			if err := p.broker.Publish(broker.PULL, broker.COORDINATOR, pull.SDP); err != nil {
				log.Printf("Failed to publish to broker: %v", err)
			}
		}
	}
}
