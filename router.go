package router

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"router/middleware"
)

// Verify interface compliance
var _ http.Handler = &HttpRouter{}

// HTTP request context key
var contextKey requestContextKey = "request-context"

type requestContextKey string

type HttpRouter struct {
	tree            *tree
	middlewareChain []middleware.Middleware
}

func NewHttpRouter() *HttpRouter {
	return &HttpRouter{
		tree:            NewTree(),
		middlewareChain: []middleware.Middleware{},
	}
}

// Configuration functions

func (r *HttpRouter) Handle(method HttpMethod, route string, handler http.Handler) *HttpRouter {
	r.tree.Register(method, route, handler)
	return r
}

func (r *HttpRouter) HandleFunc(method HttpMethod, route string, handler http.HandlerFunc) *HttpRouter {
	r.tree.Register(method, route, handler)
	return r
}

func (r *HttpRouter) UseMiddleware(middleware middleware.Middleware) *HttpRouter {
	r.middlewareChain = append(r.middlewareChain, middleware)
	return r
}

func (r *HttpRouter) UseMiddlewares(middleware ...middleware.Middleware) *HttpRouter {
	r.middlewareChain = append(r.middlewareChain, middleware...)
	return r
}

func (r *HttpRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var method HttpMethod
	if req.Method == "" {
		method = GET
	} else {
		method = HttpMethod(req.Method)
	}

	routeData, err := r.tree.Find(method, req.URL)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.NotFound(w, req)
		} else {
			panic(fmt.Sprintf("unhandled error finding request handler: %v", err))
		}

		return
	}

	// Store route data in context
	ctx := context.WithValue(req.Context(), contextKey, routeData.Context)
	reqWithContext := req.WithContext(ctx)
	*req = *reqWithContext

	// Middleware chain
	handler := middleware.GetHandlerChain(routeData.Handler, r.middlewareChain)

	// Request execution
	handler.ServeHTTP(w, reqWithContext)
}
