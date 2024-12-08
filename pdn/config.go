package pdn

import (
	"pdn/coordinator"
	"pdn/database"
	"pdn/signal"
)

type Config struct {
	Signal      signal.Config
	Database    database.Config
	Coordinator coordinator.Config
}
