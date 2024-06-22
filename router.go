package router

import (
	"context"
	"errors"
	"net/http"
)

// Verify interface compliance
var _ http.Handler = &HttpRouter{}

// HTTP request context key
var contextKey requestContextKey = "request-context"

type requestContextKey string

type HttpRouter struct {
	tree            *tree
	middlewareChain []Middleware
}

func NewHttpRouter() *HttpRouter {
	return &HttpRouter{
		tree:            NewTree(),
		middlewareChain: []Middleware{},
	}
}

// Configuration functions

func (r *HttpRouter) Handle(method HttpMethod, route string, handler http.Handler) *HttpRouter {
	r.tree.Register(method, route, handler)
	return r
}

func (r *HttpRouter) UseMiddleware(middleware Middleware) *HttpRouter {
	r.middlewareChain = append(r.middlewareChain, middleware)
	return r
}

func (r *HttpRouter) UseMiddlewares(middleware ...Middleware) *HttpRouter {
	r.middlewareChain = append(r.middlewareChain, middleware...)
	return r
}

// Start HTTP server
func (r *HttpRouter) Start(address string) error {
	return http.ListenAndServe(address, r)
}

func (r *HttpRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var method HttpMethod
	if req.Method == "" {
		method = GET
	} else {
		method = HttpMethod(req.Method)
	}

	routeData, err := r.tree.Find(method, req.URL.Path)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
		} else {
			// todo: log
		}

		return
	}

	// Store route data in context
	ctx := context.WithValue(req.Context(), contextKey, routeData.Context)
	reqWithContext := req.WithContext(ctx)
	*req = *reqWithContext

	// Middleware chain
	handler := getHandlerChain(routeData.Handler, r.middlewareChain)

	// Request execution
	handler.ServeHTTP(w, reqWithContext)
}
