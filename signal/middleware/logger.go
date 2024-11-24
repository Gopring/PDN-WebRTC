// Package middleware contains common middleware functions for HTTP handlers.
package middleware

import (
	"bufio"
	"errors"
	"log"
	"net"
	"net/http"
)

// Logger logs requests and responses.
type Logger struct {
	http.ResponseWriter
}

type logWriter struct {
	http.ResponseWriter
	statusCode int
}

// NewLogger creates a new Logger middleware.
func NewLogger() *Logger {
	return &Logger{}
}

// Hijack hijacks the connection. This is necessary for using websockets.
func (l *Logger) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := l.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("hijack not supported")
	}
	return h.Hijack()
}

// Hijack hijacks the connection. This is necessary for using websockets.
func (l *logWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := l.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("hijack not supported")
	}
	return h.Hijack()
}

func (l *logWriter) WriteHeader(code int) {
	l.statusCode = code
	l.ResponseWriter.WriteHeader(code)
}

// Intercept logs the request and response.
func (l *Logger) Intercept(next http.Handler) http.Handler {
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
