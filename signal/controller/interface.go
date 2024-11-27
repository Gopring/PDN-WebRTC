// Package controller handles HTTP logic.
package controller

import (
	"pdn/pkg/socket"
)

// Controller is an interface for handling HTTP requests.
//
//go:generate mockgen -destination=mock_controller.go -package=controller . Controller
type Controller interface {
	Process(s socket.Socket) error
}
