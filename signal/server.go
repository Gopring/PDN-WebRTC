// Package signal contains business logic and prerequisites for WebRTC streaming.
package signal

import (
	"fmt"
	"log"
	"net/http"
	"pdn/broker"
	"pdn/media"
	"pdn/metric"
	"pdn/signal/controller"
	"pdn/signal/handler"
	"time"
)

// Signal contains the server and configuration.
type Signal struct {
	server  *http.Server
	conf    Config
	metrics *metric.Metrics
}

// New creates a new instance of Signal.
func New(config Config, metrics *metric.Metrics) *Signal {
	brk := broker.New()
	con := controller.New(brk)
	med := media.New(brk, metrics)
	go med.Run()
	srv := &http.Server{
		Addr:        fmt.Sprintf(":%d", config.Port),
		ReadTimeout: 2 * time.Second,
		Handler:     handler.New(con, metrics),
	}
	return &Signal{
		server:  srv,
		conf:    config,
		metrics: metrics,
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
