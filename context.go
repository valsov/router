package router

// Context of the request, contains URL parameters
type RequestContext struct {
	RouteParams map[string]string
	QueryParams map[string][]string
}

// Retrieve a parameter value from the request route
func (ctx *RequestContext) GetRouteParam(param string) string {
	return ctx.RouteParams[param]
}

// Retrieve the first value of a parameter from the request query
func (ctx *RequestContext) GetQueryParam(param string) string {
	if values, found := ctx.QueryParams[param]; found && len(values) != 0 {
		return values[0]
	}
	return ""
}

// Retrieve all values of a parameter from the request query
func (ctx *RequestContext) GetQueryParamValues(param string) []string {
	return ctx.QueryParams[param]
}

// Retrieve a parameter value from the request. Both route and query are searched.
func (ctx *RequestContext) GetParam(param string) string {
	routeParam := ctx.GetRouteParam(param)
	if routeParam != "" {
		return routeParam
	}

	return ctx.GetQueryParam(param)
}
