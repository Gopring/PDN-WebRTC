// Package middleware contains common middleware functions for HTTP handlers.
package middleware

import "net/http"

// CORS sets up CORS headers.
type CORS struct {
}

// NewCORS creates a new CORS middleware.
func NewCORS() *CORS {
	return &CORS{}
}

// Intercept sets up CORS headers.
func (c CORS) Intercept(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, channel-key")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
