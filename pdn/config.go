// Package pdn is middleware that Peer-assisted Delivery Network with WebRTC.
package pdn

import (
	"pdn/coordinator"
	"pdn/database"
	"pdn/media"
	"pdn/metric"
	"pdn/signal"
)

// Config contains the configuration for the PDN.
type Config struct {
	Signal      signal.Config
	Database    database.Config
	Coordinator coordinator.Config
	Metrics     metric.Config
	Media       media.Config
}
