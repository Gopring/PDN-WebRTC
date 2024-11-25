// Package signal contains business logic and prerequisites for WebRTC streaming.
package signal

import (
	"fmt"
	"log"
	"net/http"
	"pdn/media"
	"pdn/signal/controller"
	"pdn/signal/coordinator"
	"pdn/signal/middleware"
	"time"
)

// Signal contains the server and configuration.
type Signal struct {
	server *http.Server
	conf   Config
}

// New creates a new instance of Signal.
func New(config Config) *Signal {

	med := media.New()
	cod := coordinator.New(med)
	con := controller.New(cod, config.Debug)
	mds := []middleware.Interceptor{
		middleware.NewAuth(),
		middleware.NewLogger(),
		middleware.NewSocket(con),
	}
	mux := middleware.Set(http.NewServeMux(), mds...)

	srv := &http.Server{
		Addr:        fmt.Sprintf(":%d", config.Port),
		ReadTimeout: 2 * time.Second,
		Handler:     mux,
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
