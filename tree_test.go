package router

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"testing"
)

func TestRegister(t *testing.T) {
	testCases := []struct {
		method      HttpMethod
		route       string
		shouldPanic bool
	}{
		{
			method:      GET,
			route:       "/",
			shouldPanic: false,
		},
		{
			method:      GET,
			route:       "/test",
			shouldPanic: false,
		},
		{
			method:      GET,
			route:       "/test", // Already registered
			shouldPanic: true,
		},
		{
			method:      "UNKNOWN_METHOD",
			route:       "/",
			shouldPanic: true,
		},
		{
			method:      GET,
			route:       "/{wild}/test/{wild}", // Duplicated wildcard parameter name
			shouldPanic: true,
		},
	}

	tree := NewTree()
	handler := http.NotFoundHandler() // Sample handler
	for _, tc := range testCases {
		testFn := func() {
			tree.Register(tc.method, tc.route, handler)
		}
		err := checkPanic(testFn, tc.shouldPanic)
		if err != nil {
			t.Errorf("%v", err)
		}
	}
}

func TestFind(t *testing.T) {
	testCases := []struct {
		method      HttpMethod
		url         *url.URL
		urlParams   map[string]string
		queryParams map[string][]string
		expectedErr error
	}{
		{
			method:      "UNKNOWN_METHOD",
			url:         getUrl("/"),
			expectedErr: ErrUnhandledMethod,
		},
		{
			method:      GET,
			url:         getUrl("/notfound"),
			expectedErr: ErrNotFound,
		},
		{
			method: GET,
			url:    getUrl("/"),
		},
		{
			method: GET,
			url:    getUrl("/test1/test2/test3"),
		},
		{
			method:      GET,
			url:         getUrl("/test1/test2/test3?queryparam=test1&queryparam=test2"),
			queryParams: map[string][]string{"queryparam": {"test1", "test2"}},
		},
		{
			method:    GET,
			url:       getUrl("/test1/test2/test3/wildvalue2"), // Favor non-wildcard routes
			urlParams: map[string]string{"wild2": "wildvalue2"},
		},
		{
			method: GET,
			url:    getUrl("/test1/wildvalue1/test3/wildvalue2"),
			urlParams: map[string]string{
				"wild1": "wildvalue1",
				"wild2": "wildvalue2",
			},
		},
	}

	handler := http.NotFoundHandler() // Sample handler
	tree := NewTree()
	tree.Register(GET, "/", handler)
	tree.Register(GET, "/test1/test2/test3", handler)
	tree.Register(GET, "/test1/{wild1}/test3/{wild2}", handler)
	tree.Register(GET, "/test1/test2/test3/{wild2}", handler)

	for _, tc := range testCases {
		routeData, err := tree.Find(tc.method, tc.url)
		if tc.expectedErr != nil {
			if err == nil {
				t.Errorf("expected error=%v, got none", tc.expectedErr)
			} else if !errors.Is(err, tc.expectedErr) {
				t.Errorf("expected error=%v, got=%v", tc.expectedErr, err)
			}
		} else if err != nil {
			t.Errorf("expected error: %v", err)
		} else {
			if tc.urlParams != nil {
				if len(tc.urlParams) != len(routeData.Context.RouteParams) {
					t.Errorf("got wrong url parameters count. expected=%v, got=%v", len(tc.urlParams), len(routeData.Context.RouteParams))
				}
				for paramKey, paramVal := range routeData.Context.RouteParams {
					if _, found := tc.urlParams[paramKey]; !found {
						t.Errorf("url parameter not found: %s", paramKey)
					} else if paramVal != tc.urlParams[paramKey] {
						t.Errorf("url parameter doesn't match the expected (%s). expected=%s, got=%s", paramKey, tc.urlParams[paramKey], paramVal)
					}
				}
			}
			if tc.queryParams != nil {
				if len(tc.queryParams) != len(routeData.Context.QueryParams) {
					t.Errorf("got wrong query parameters count. expected=%v, got=%v", len(tc.queryParams), len(routeData.Context.QueryParams))
				}
				for paramKey, paramVal := range routeData.Context.QueryParams {
					if _, found := tc.queryParams[paramKey]; !found {
						t.Errorf("query parameter not found: %s", paramKey)
					} else if len(paramVal) != len(tc.queryParams[paramKey]) {
						t.Errorf("query parameter %s count doesn't match the expected. expected=%d, got=%d", paramKey, len(tc.queryParams[paramKey]), len(paramVal))
					}

					for i := 0; i < len(paramVal); i++ {
						if paramVal[i] != tc.queryParams[paramKey][i] {
							t.Errorf("query parameter value doesn't match the expected (%s). expected=%s, got=%s", paramKey, tc.queryParams[paramKey][i], paramVal[i])
						}
					}
				}
			}
		}
	}
}

func getUrl(route string) *url.URL {
	result, _ := url.ParseRequestURI(fmt.Sprintf("https://127.0.0.1%s", route))
	return result
}

func checkPanic(testFn func(), shouldPanic bool) (err error) {
	defer func() {
		r := recover()
		if (r == nil) == shouldPanic {
			// Unexpected panic state
			var expectedDetail string
			if shouldPanic {
				expectedDetail = "a panic was expected but didn't occur"
			} else {
				expectedDetail = "a panic was unexpected but occurred"
			}
			err = fmt.Errorf("panic check failed: %s", expectedDetail)
		}
	}()

	testFn()
	return err
}
