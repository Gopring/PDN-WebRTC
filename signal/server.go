// Package signal contains business logic and prerequisites for WebRTC streaming.
package signal

import (
	"fmt"
	"log"
	"net/http"
	"pdn/broker"
	"pdn/signal/controller"
	"time"
)

// Signal contains the server and configuration.
type Signal struct {
	server *http.Server
	conf   Config
}

// New creates a new instance of Signal.
func New(config Config) *Signal {
	brk := broker.New()
	con := controller.New(brk)

	srv := &http.Server{
		Addr:        fmt.Sprintf(":%d", config.Port),
		ReadTimeout: 2 * time.Second,
		Handler:     con,
	}
	return &Signal{
		server: srv,
		conf:   config,
	}
}

// Start runs the signal server.
func (s *Signal) Start() error {
	if s.conf.CertFile == "" || s.conf.KeyFile == "" {
		log.Printf("Starting server port on %d, without TLS", s.conf.Port)
		if err := s.server.ListenAndServe(); err != nil {
			return fmt.Errorf("failed to start server: %w", err)
		}
		return nil
	}

	log.Printf("Starting server port on %d, with TLS", s.conf.Port)
	if err := s.server.ListenAndServeTLS(s.conf.CertFile, s.conf.KeyFile); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}
