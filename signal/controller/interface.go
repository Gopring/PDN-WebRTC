// Package controller handles HTTP logic.
package controller

import "net/http"

// Controller is an interface for handling HTTP requests.
//
//go:generate mockgen -destination=mock_controller.go -package=controller . Controller
type Controller interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}
