// Package pdn is middleware that Peer-assisted Delivery Network with WebRTC.
package pdn

import (
	"fmt"
	"pdn/broker"
	"pdn/coordinator"
	"pdn/database"
	"pdn/database/memory"
	"pdn/media"
	"pdn/metric"
	"pdn/signal"
)

// PDN contains servers and configuration.
type PDN struct {
	broker      *broker.Broker
	database    database.Database
	media       *media.Media
	coordinator *coordinator.Coordinator
	signal      *signal.Signal
	metric      *metric.Metrics
}

// New creates a new instance of PDN.
func New(config Config) *PDN {
	brk := broker.New()
	db := memory.New(config.Database)
	med := media.New(brk)
	cod := coordinator.New(config.Coordinator, brk, db)
	met := metric.New(config.Metrics)
	sig := signal.New(config.Signal, db, brk, met)

	return &PDN{
		broker:      brk,
		database:    db,
		media:       med,
		coordinator: cod,
		signal:      sig,
		metric:      met,
	}
}

// Start runs the signal server and metrics server.
func (p *PDN) Start() error {

	go p.metric.Start()
	go p.media.Start()
	go p.coordinator.Start()
	if err := p.signal.Start(); err != nil {
		return fmt.Errorf("failed to start signal server: %w", err)
	}
	return nil
}
