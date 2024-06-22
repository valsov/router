package router

import "net/http"

type Middleware func(http.Handler) http.Handler

func getHandlerChain(handler http.Handler, middleware []Middleware) http.Handler {
	if len(middleware) == 0 {
		return handler
	}

	currentHandler := middleware[len(middleware)-1](handler) // Add final (endpoint) HTTP handler
	for i := len(middleware) - 2; i >= 0; i-- {
		currentHandler = middleware[i](currentHandler)
	}

	return currentHandler
}

/*func sample() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}
}*/
