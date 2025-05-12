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
	"pdn/pool"
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
	pool        *pool.Pool
}

// New creates a new instance of PDN.
func New(config Config) *PDN {
	met := metric.New(config.Metrics)
	brk := broker.New()
	db := memory.New(config.Database)
	med := media.New(config.Media, brk, met)
	pl := pool.New(db)
	cod := coordinator.New(config.Coordinator, brk, met, db, pl)
	sig := signal.New(config.Signal, db, brk, met)

	return &PDN{
		broker:      brk,
		database:    db,
		media:       med,
		pool:        pl,
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
