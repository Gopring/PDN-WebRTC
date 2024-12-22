// Package cmd parse args to configure application.
package cmd

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"pdn/coordinator"
	"pdn/database"
	"pdn/metric"
	"pdn/pdn"
	"pdn/signal"
)

// Run starts the application.
func Run() {
	config, err := SetupConfig(os.Stdout, os.Args[1:])
	if err != nil {
		os.Exit(1)
	}

	p := pdn.New(config)
	if err = p.Start(); err != nil {
		os.Exit(1)
	}
}

// SetupConfig sets up and returns the configuration.
func SetupConfig(w io.Writer, args []string) (pdn.Config, error) {
	config, err := Parse(w, args)
	if err != nil {
		return config, err
	}
	if err = config.Signal.Validate(); err != nil {
		return config, err
	}
	return config, nil
}

// Parse parses the command line arguments.
func Parse(w io.Writer, args []string) (pdn.Config, error) {
	sig := signal.Config{}
	db := database.Config{}
	cor := coordinator.Config{}
	met := metric.Config{}
	fs := flag.NewFlagSet("config", flag.ContinueOnError)
	fs.SetOutput(w)
	fs.IntVar(&sig.Port, "port", signal.DefaultPort, "listening port")
	fs.BoolVar(&sig.Debug, "debug", false, "debug mode")
	fs.StringVar(&sig.KeyFile, "key", "", "key file path")
	fs.StringVar(&sig.CertFile, "cert", "", "cert file path")
	fs.BoolVar(&db.SetDefaultChannel, "setDefaultChannel", false, "set default channel for debug or test")
	fs.IntVar(&cor.MaxForwardingNumber, "maxForwardingNumber",
		coordinator.DefaultMaxForwardingNumber, "max forwarding number")
	fs.BoolVar(&cor.SetPeerConnection, "setPeerConnection",
		coordinator.DefaultSetPeerConnection, "set peer assisted delivery network mode")
	fs.IntVar(&met.Port, "metricPort", metric.DefaultMetricsPort, "listening port")
	fs.StringVar(&met.Path, "metricPath", metric.DefaultMetricsPath, "metrics path")

	err := fs.Parse(args)
	if err != nil {
		return pdn.Config{}, fmt.Errorf("failed to parse args: %w", err)
	}

	if fs.NArg() != 0 {
		return pdn.Config{}, errors.New("some args are not parsed")
	}

	return pdn.Config{
		Signal:      sig,
		Database:    db,
		Coordinator: cor,
		Metrics:     met,
	}, nil
}
