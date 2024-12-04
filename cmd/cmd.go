// Package cmd parse args to configure application.
package cmd

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"pdn/metric"
	"pdn/signal"
)

// Run starts the application.
func Run() {
	config, err := SetupConfig(os.Stdout, os.Args[1:])
	if err != nil {
		os.Exit(1)
	}

	// Setup metrics configuration
	metrics := metric.New(metric.Config{
		Port: metric.DefaultMetricsPort,
		Path: metric.DefaultMetricsPath,
	})
	metrics.RegisterMetrics()
	stop := make(chan struct{})
	go metrics.Start(stop)

	s := signal.New(config, metrics)
	if err = s.Start(); err != nil {
		os.Exit(1)
	}
}

// SetupConfig sets up and returns the configuration.
func SetupConfig(w io.Writer, args []string) (signal.Config, error) {
	config, err := Parse(w, args)
	if err != nil {
		return config, err
	}
	if err = config.Validate(); err != nil {
		return config, err
	}
	return config, nil
}

// Parse parses the command line arguments.
func Parse(w io.Writer, args []string) (signal.Config, error) {
	con := signal.Config{}

	fs := flag.NewFlagSet("pdn", flag.ContinueOnError)
	fs.SetOutput(w)
	fs.IntVar(&con.Port, "port", signal.DefaultPort, "listening port")
	fs.BoolVar(&con.Debug, "debug", false, "debug mode")
	fs.StringVar(&con.KeyFile, "key", "", "key file path")
	fs.StringVar(&con.CertFile, "cert", "", "cert file path")

	err := fs.Parse(args)
	if err != nil {
		return signal.Config{}, fmt.Errorf("failed to parse args: %w", err)
	}

	if fs.NArg() != 0 {
		return signal.Config{}, errors.New("some args are not parsed")
	}

	return con, nil
}
