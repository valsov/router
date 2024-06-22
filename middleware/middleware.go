package middleware

import "net/http"

// Middleware function, accepting a next http.Handler as parameter
type Middleware func(http.Handler) http.Handler

// Produce a `http.Handler` that contains the whole middleware chain (in the given order) and the final http.Handler at the end
func GetHandlerChain(handler http.Handler, middleware []Middleware) http.Handler {
	if len(middleware) == 0 {
		return handler
	}

	currentHandler := middleware[len(middleware)-1](handler) // Add final (endpoint) HTTP handler
	for i := len(middleware) - 2; i >= 0; i-- {
		currentHandler = middleware[i](currentHandler)
	}

	return currentHandler
}
