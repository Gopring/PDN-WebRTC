package pdn

import (
	"fmt"
	"pdn/broker"
	"pdn/coordinator"
	"pdn/database"
	"pdn/database/memory"
	"pdn/media"
	"pdn/signal"
)

// PDN contains servers and configuration.
type PDN struct {
	broker      *broker.Broker
	database    database.Database
	media       *media.Media
	coordinator *coordinator.Coordinator
	signal      *signal.Signal
}

// New creates a new instance of PDN.
func New(config Config) *PDN {
	brk := broker.New()
	db := memory.New(config.Database)
	med := media.New(brk)
	cod := coordinator.New(config.Coordinator, brk, db)
	sig := signal.New(config.Signal, db, brk)
	return &PDN{
		broker:      brk,
		database:    db,
		media:       med,
		coordinator: cod,
		signal:      sig,
	}
}

// Start runs the signal server.
func (p *PDN) Start() error {
	go p.media.Start()
	go p.coordinator.Start()
	if err := p.signal.Start(); err != nil {
		return fmt.Errorf("failed to start signal server: %w", err)
	}
	return nil
}
