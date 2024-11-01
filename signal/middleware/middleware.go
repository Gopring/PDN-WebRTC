// Package middleware contains common middleware functions for HTTP handlers.
package middleware

import "net/http"

// Interceptor is a middleware interface.
type Interceptor interface {
	Intercept(handlerFunc http.Handler) http.Handler
}

// Set applies multiple middleware to a handler. The middleware are applied in
// the order they are passed. For example: if Set(auth, cors, logger) is called,
// the request will first be logged, then CORS headers will be set, and finally
func Set(h http.Handler, m ...Interceptor) http.Handler {
	for _, i := range m {
		h = i.Intercept(h)
	}
	return h
}
