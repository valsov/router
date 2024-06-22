package router

import "testing"

func TestGetRouteParam(t *testing.T) {
	testCases := []struct {
		routeParams map[string]string
		queryParams map[string]string
		search      string
		expected    string
	}{
		{
			// Exists
			routeParams: map[string]string{"p1": "propVal1", "p2": "propVal2"},
			queryParams: map[string]string{"p2": "propVal2FromQuery"},
			search:      "p2",
			expected:    "propVal2",
		},
		{
			// Exists in query, but not in route
			routeParams: map[string]string{"p1": "propVal1"},
			queryParams: map[string]string{"p2": "propVal2FromQuery"},
			search:      "p2",
			expected:    "",
		},
	}

	for _, tc := range testCases {
		reqCtx := RequestContext{
			RouteParams: tc.routeParams,
			QueryParams: tc.queryParams,
		}

		result := reqCtx.GetRouteParam(tc.search)
		if result != tc.expected {
			t.Errorf("search yield unexpected result. expected=%s, got=%s", tc.expected, result)
		}
	}
}

func TestGetQueryParam(t *testing.T) {
	testCases := []struct {
		routeParams map[string]string
		queryParams map[string]string
		search      string
		expected    string
	}{
		{
			// Exists
			routeParams: map[string]string{"p2": "propVal2FromRoute"},
			queryParams: map[string]string{"p1": "propVal1", "p2": "propVal2"},
			search:      "p2",
			expected:    "propVal2",
		},
		{
			// Exists in route, but not in query
			routeParams: map[string]string{"p2": "propVal2FromRoute"},
			queryParams: map[string]string{"p1": "propVal1"},
			search:      "p2",
			expected:    "",
		},
	}

	for _, tc := range testCases {
		reqCtx := RequestContext{
			RouteParams: tc.routeParams,
			QueryParams: tc.queryParams,
		}

		result := reqCtx.GetQueryParam(tc.search)
		if result != tc.expected {
			t.Errorf("search yield unexpected result. expected=%s, got=%s", tc.expected, result)
		}
	}
}

func TestGetParam(t *testing.T) {
	testCases := []struct {
		routeParams map[string]string
		queryParams map[string]string
		search      string
		expected    string
	}{
		{
			// Exists in route
			routeParams: map[string]string{"p1": "propVal1", "p2": "propVal2"},
			queryParams: map[string]string{"p3": "propVal3"},
			search:      "p1",
			expected:    "propVal1",
		},
		{
			// Exists in query
			routeParams: map[string]string{"p1": "propVal1", "p2": "propVal2"},
			queryParams: map[string]string{"p3": "propVal3"},
			search:      "p3",
			expected:    "propVal3",
		},
		{
			// Exists in both route and query (favors route)
			routeParams: map[string]string{"p1": "propVal1", "p2": "propVal2FromRoute"},
			queryParams: map[string]string{"p2": "propVal2FromQuery"},
			search:      "p2",
			expected:    "propVal2FromRoute",
		},
		{
			// Not found
			routeParams: map[string]string{"p1": "propVal1"},
			queryParams: map[string]string{"p2": "propVal2"},
			search:      "p3",
			expected:    "",
		},
	}

	for _, tc := range testCases {
		reqCtx := RequestContext{
			RouteParams: tc.routeParams,
			QueryParams: tc.queryParams,
		}

		result := reqCtx.GetParam(tc.search)
		if result != tc.expected {
			t.Errorf("search yield unexpected result. expected=%s, got=%s", tc.expected, result)
		}
	}
}
