package router

import (
	"errors"
	"fmt"
	"net/http"
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
		route       string
		urlParams   map[string]string
		expectedErr error
	}{
		{
			method:      "UNKNOWN_METHOD",
			route:       "/",
			expectedErr: ErrUnhandledMethod,
		},
		{
			method:      GET,
			route:       "/notfound",
			expectedErr: ErrNotFound,
		},
		{
			method: GET,
			route:  "/",
		},
		{
			method: GET,
			route:  "/test1/test2/test3",
		},
		{
			method:    GET,
			route:     "/test1/test2/test3/wildvalue2", // Favor non-wildcard routes
			urlParams: map[string]string{"wild2": "wildvalue2"},
		},
		{
			method: GET,
			route:  "/test1/wildvalue1/test3/wildvalue2",
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
		routeData, err := tree.Find(tc.method, tc.route)
		if tc.expectedErr != nil {
			if err == nil {
				t.Errorf("expected error=%v, got none", tc.expectedErr)
			} else if !errors.Is(err, tc.expectedErr) {
				t.Errorf("expected error=%v, got=%v", tc.expectedErr, err)
			}
		} else if err != nil {
			t.Errorf("expected error: %v", err)
		} else if tc.urlParams != nil {
			if len(tc.urlParams) != len(routeData.Context.RouteParams) {
				t.Errorf("got wrong url parameters count. expected=%v, got=%v", len(tc.urlParams), len(routeData.Context.RouteParams))
			}
			for paramKey, paramVal := range routeData.Context.RouteParams {
				if _, found := tc.urlParams[paramKey]; !found {
					t.Errorf("url parameter not found: %s", paramKey)
				} else if paramVal != tc.urlParams[paramKey] {
					t.Errorf("url parameter doesn't match the expected (%s) expected=%s, got=%s", paramKey, tc.urlParams[paramKey], paramVal)
				}
			}
		}
	}
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
