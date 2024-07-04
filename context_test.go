package router

import (
	"net/http"
	"testing"
)

func TestGetRouteParam(t *testing.T) {
	testCases := []struct {
		routeParams map[string]string
		queryParams map[string][]string
		search      string
		expected    string
		found       bool
	}{
		{
			// Exists
			routeParams: map[string]string{"p1": "propVal1", "p2": "propVal2"},
			queryParams: map[string][]string{"p2": {"propVal2FromQuery"}},
			search:      "p2",
			expected:    "propVal2",
			found:       true,
		},
		{
			// Exists in query, but not in route
			routeParams: map[string]string{"p1": "propVal1"},
			queryParams: map[string][]string{"p2": {"propVal2FromQuery"}},
			search:      "p2",
			expected:    "",
			found:       false,
		},
	}

	for _, tc := range testCases {
		reqCtx := requestContext{
			RouteParams: tc.routeParams,
			QueryParams: tc.queryParams,
		}
		r := buildRequestWithContext(reqCtx)

		result, found := GetRouteParam(r, tc.search)
		if found != tc.found {
			t.Errorf("invalid search status result. expected=%t, got=%t", tc.found, found)
		}
		if result != tc.expected {
			t.Errorf("search yield unexpected result. expected=%s, got=%s", tc.expected, result)
		}
	}
}

func TestGetQueryParam(t *testing.T) {
	testCases := []struct {
		routeParams map[string]string
		queryParams map[string][]string
		search      string
		expected    string
		found       bool
	}{
		{
			// Exists
			routeParams: map[string]string{"p2": "propVal2FromRoute"},
			queryParams: map[string][]string{"p1": {"propVal1"}, "p2": {"propVal2"}},
			search:      "p2",
			expected:    "propVal2",
			found:       true,
		},
		{
			// Exists with 2 values, get first one
			routeParams: map[string]string{"p2": "propVal2FromRoute"},
			queryParams: map[string][]string{"p1": {"propVal1"}, "p2": {"propVal2", "propVal3"}},
			search:      "p2",
			expected:    "propVal2",
			found:       true,
		},
		{
			// Exists in route, but not in query
			routeParams: map[string]string{"p2": "propVal2FromRoute"},
			queryParams: map[string][]string{"p1": {"propVal1"}},
			search:      "p2",
			expected:    "",
			found:       false,
		},
	}

	for _, tc := range testCases {
		reqCtx := requestContext{
			RouteParams: tc.routeParams,
			QueryParams: tc.queryParams,
		}
		r := buildRequestWithContext(reqCtx)

		result, found := GetQueryParam(r, tc.search)
		if found != tc.found {
			t.Errorf("invalid search status result. expected=%t, got=%t", tc.found, found)
		}
		if result != tc.expected {
			t.Errorf("search yield unexpected result. expected=%s, got=%s", tc.expected, result)
		}
	}
}

func TestGetQueryParamValues(t *testing.T) {
	testCases := []struct {
		routeParams map[string]string
		queryParams map[string][]string
		search      string
		expected    []string
		found       bool
	}{
		{
			// Exists
			routeParams: map[string]string{"p2": "propVal2FromRoute"},
			queryParams: map[string][]string{"p1": {"propVal1"}, "p2": {"propVal2"}},
			search:      "p2",
			expected:    []string{"propVal2"},
			found:       true,
		},
		{
			// Exists with 2 values, get all
			routeParams: map[string]string{"p2": "propVal2FromRoute"},
			queryParams: map[string][]string{"p1": {"propVal1"}, "p2": {"propVal2", "propVal3"}},
			search:      "p2",
			expected:    []string{"propVal2", "propVal3"},
			found:       true,
		},
		{
			// Exists in route, but not in query
			routeParams: map[string]string{"p2": "propVal2FromRoute"},
			queryParams: map[string][]string{"p1": {"propVal1"}},
			search:      "p2",
			expected:    nil,
			found:       false,
		},
	}

	for _, tc := range testCases {
		reqCtx := requestContext{
			RouteParams: tc.routeParams,
			QueryParams: tc.queryParams,
		}
		r := buildRequestWithContext(reqCtx)

		result, found := GetQueryParamValues(r, tc.search)
		if found != tc.found {
			t.Errorf("invalid search status result. expected=%t, got=%t", tc.found, found)
		}
		if len(result) != len(tc.expected) {
			t.Errorf("search yield unexpected result length. expected=%d, got=%d", len(tc.expected), len(result))
		} else {
			for i := 0; i < len(result); i++ {
				if tc.expected[i] != result[i] {
					t.Errorf("search yield unexpected result at index [%d]. expected=%s, got=%s", i, tc.expected[i], result[i])
				}
			}
		}
	}
}

func TestGetParam(t *testing.T) {
	testCases := []struct {
		routeParams map[string]string
		queryParams map[string][]string
		search      string
		expected    string
		found       bool
	}{
		{
			// Exists in route
			routeParams: map[string]string{"p1": "propVal1", "p2": "propVal2"},
			queryParams: map[string][]string{"p3": {"propVal3"}},
			search:      "p1",
			expected:    "propVal1",
			found:       true,
		},
		{
			// Exists in query
			routeParams: map[string]string{"p1": "propVal1", "p2": "propVal2"},
			queryParams: map[string][]string{"p3": {"propVal3"}},
			search:      "p3",
			expected:    "propVal3",
			found:       true,
		},
		{
			// Exists in both route and query (favors route)
			routeParams: map[string]string{"p1": "propVal1", "p2": "propVal2FromRoute"},
			queryParams: map[string][]string{"p2": {"propVal2FromQuery"}},
			search:      "p2",
			expected:    "propVal2FromRoute",
			found:       true,
		},
		{
			// Not found
			routeParams: map[string]string{"p1": "propVal1"},
			queryParams: map[string][]string{"p2": {"propVal2"}},
			search:      "p3",
			expected:    "",
			found:       false,
		},
	}

	for _, tc := range testCases {
		reqCtx := requestContext{
			RouteParams: tc.routeParams,
			QueryParams: tc.queryParams,
		}
		r := buildRequestWithContext(reqCtx)

		result, found := GetParam(r, tc.search)
		if found != tc.found {
			t.Errorf("invalid search status result. expected=%t, got=%t", tc.found, found)
		}
		if result != tc.expected {
			t.Errorf("search yield unexpected result. expected=%s, got=%s", tc.expected, result)
		}
	}
}

func buildRequestWithContext(rCtx requestContext) *http.Request {
	r := http.Request{}
	return newRequestWithContext(&r, rCtx)
}
