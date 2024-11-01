// Package middleware contains common middleware functions for HTTP handlers.
package middleware

import "net/http"

// Auth validates the channel key and channel id.
type Auth struct {
	//TODO(window9u): we need to add a config, memory storage client
}

// NewAuth creates a new Auth middleware.
func NewAuth() *Auth {
	return &Auth{}
}

// Intercept processes the request and call the next handler.
func (a Auth) Intercept(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//TODO(window9u): we need to validate channel key and channel id
		// and insert it into request context

		next.ServeHTTP(w, r)
	})
}
