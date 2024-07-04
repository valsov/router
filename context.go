package router

import (
	"context"
	"net/http"
)

// HTTP request context key
var contextKey requestContextKey = "request-context"

type requestContextKey string

// Context of the request, contains URL parameters
type requestContext struct {
	RouteParams map[string]string
	QueryParams map[string][]string
}

// Retrieve a parameter value from the request route
func GetRouteParam(r *http.Request, param string) (string, bool) {
	ctx, found := getRequestContext(r)
	if !found {
		return "", false
	}
	val, found := ctx.RouteParams[param]
	return val, found
}

// Retrieve the first value of a parameter from the request query
func GetQueryParam(r *http.Request, param string) (string, bool) {
	ctx, found := getRequestContext(r)
	if !found {
		return "", false
	}
	if values, found := ctx.QueryParams[param]; found && len(values) != 0 {
		return values[0], true
	}
	return "", false
}

// Retrieve all values of a parameter from the request query
func GetQueryParamValues(r *http.Request, param string) ([]string, bool) {
	ctx, found := getRequestContext(r)
	if !found {
		return nil, false
	}
	val, found := ctx.QueryParams[param]
	return val, found
}

// Retrieve a parameter value from the request. Both route and query are searched.
func GetParam(r *http.Request, param string) (string, bool) {
	routeParam, found := GetRouteParam(r, param)
	if found {
		return routeParam, true
	}

	queryParam, found := GetQueryParam(r, param)
	return queryParam, found
}

// Try to extract a requestContext from the given request
func getRequestContext(r *http.Request) (requestContext, bool) {
	ctxVal := r.Context().Value(contextKey)
	var rCtx requestContext
	if ctxVal == nil {
		return rCtx, false
	}

	return ctxVal.(requestContext), true
}

// Produce a new request with the given requestContext injected into its context
func newRequestWithContext(r *http.Request, rCtx requestContext) *http.Request {
	ctx := context.WithValue(r.Context(), contextKey, rCtx)
	return r.WithContext(ctx)
}
