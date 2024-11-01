// Package middleware contains common middleware functions for HTTP handlers.
package middleware

import (
	"log"
	"net/http"
)

// Logger logs requests and responses.
type Logger struct {
}

type logWriter struct {
	http.ResponseWriter
	statusCode int
}

func (l *logWriter) WriteHeader(code int) {
	l.statusCode = code
	l.ResponseWriter.WriteHeader(code)
}

// NewLogger creates a new Logger middleware.
func NewLogger() *Logger {
	return &Logger{}
}

// Intercept logs the request and response.
func (l Logger) Intercept(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//TODO(window9u): we could export metrics here status code, response time, etc..
		// start timestamp here
		rw := logWriter{ResponseWriter: w}
		next.ServeHTTP(&rw, r)
		// end timestamp here and export to metric
		if rw.statusCode >= 400 {
			// export status code too
			log.Printf("request fails with %d", rw.statusCode)
		} else {
			log.Printf("request succeed with %d", rw.statusCode)
		}
	})
}
