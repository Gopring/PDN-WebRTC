// Package middleware contains common middleware functions for HTTP handlers.
package middleware

import (
	"bufio"
	"errors"
	"net"
	"net/http"
)

// CORS sets up CORS headers.
type CORS struct {
	http.ResponseWriter
}

// NewCORS creates a new CORS middleware.
func NewCORS() *CORS {
	return &CORS{}
}

// Hijack hijacks the connection. This is necessary for using websockets.
func (a *CORS) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := a.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("hijack not supported")
	}
	return h.Hijack()
}

// Intercept sets up CORS headers.
func (a *CORS) Intercept(next http.Handler) http.Handler {
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
