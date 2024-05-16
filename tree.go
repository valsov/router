package router

import (
	"errors"
	"fmt"
	"strings"
)

const (
	GET    HttpMethod = "GET"
	POST   HttpMethod = "POST"
	PUT    HttpMethod = "PUT"
	PATCH  HttpMethod = "PATCH"
	DELETE HttpMethod = "DELETE"

	WILDCARDCHAR byte = '{'
)

var (
	ErrUnhandledMethod error = errors.New("unhandled method")
	ErrNotFound        error = errors.New("not found")
)

var splitFn = func(c rune) bool {
	return c == '/'
}

type HttpMethod string

type RouteHandler func() // todo: input params and return type

type RouteData struct {
	Handler   func()
	UrlParams map[string]string
}

type tree struct {
	nodes [5]treeNode
}

func New() tree {
	return tree{
		[5]treeNode{
			{Content: string(GET), Children: []*treeNode{}},
			{Content: string(POST), Children: []*treeNode{}},
			{Content: string(PUT), Children: []*treeNode{}},
			{Content: string(PATCH), Children: []*treeNode{}},
			{Content: string(DELETE), Children: []*treeNode{}},
		},
	}
}

func (t *tree) GetRootNode(method HttpMethod) (*treeNode, bool) {
	switch method {
	case GET:
		return &t.nodes[0], true
	case POST:
		return &t.nodes[1], true
	case PUT:
		return &t.nodes[2], true
	case PATCH:
		return &t.nodes[3], true
	case DELETE:
		return &t.nodes[4], true
	}

	return nil, false
}

// Can panic
func (t *tree) Add(method HttpMethod, route string, handler RouteHandler) {
	root, found := t.GetRootNode(method)
	if !found {
		panic(fmt.Sprintf("%s HTTP method is not supported", method))
	}

	routeSplit := strings.FieldsFunc(route, splitFn)
	if len(routeSplit) == 0 {
		// Root path
		if root.Handler == nil {
			root.Handler = handler
			return
		} else {
			panic(fmt.Sprintf("[%s] %s was already registered with another handler", method, route))
		}
	}

	root.Add(routeSplit, 0, handler)
}

func (t *tree) Find(method HttpMethod, route string) (RouteData, error) {
	root, found := t.GetRootNode(method)
	if !found {
		return RouteData{}, ErrUnhandledMethod
	}

	routeSplit := strings.FieldsFunc(route, splitFn)
	if len(routeSplit) == 0 {
		// Root path
		if root.Handler != nil {
			return RouteData{Handler: root.Handler}, nil
		} else {
			return RouteData{}, ErrNotFound
		}
	}

	urlParams := map[string]string{}
	node, found := root.Find(routeSplit, 0, urlParams)
	if !found {
		return RouteData{}, ErrNotFound
	}

	return RouteData{node.Handler, urlParams}, nil
}

type treeNode struct {
	Content       string
	Handler       RouteHandler
	Children      []*treeNode // todo: Can be optimized, + ((??? : can have multiple wildcards name (having different children) ))
	WildCardChild *treeNode
}

// Can panic
func (node *treeNode) Add(route []string, currentIndex int, handler RouteHandler) {
	// todo: panic on duplicated urlparams

	//wildCard := route[currentIndex][0] == WILDCARDCHAR
	//if wildCard

}

func (node *treeNode) Find(route []string, currentIndex int, urlParams map[string]string) (*treeNode, bool) {
	if currentIndex == len(route)-1 {
		// Arrived at last index, try to find a handler
		if node.Handler != nil {
			return node, true
		} else {
			return nil, false
		}
	}

	for _, n := range node.Children {
		if n.Content == route[currentIndex] {
			foundNode, found := n.Find(route, currentIndex+1, urlParams)
			if found {
				return foundNode, true
			}
			break
		}
	}

	// Not found in exact children, try wildcard
	if node.WildCardChild == nil {
		return nil, false
	}
	// Populate url parameters
	urlParams[node.WildCardChild.Content] = route[currentIndex]

	// Try find complete path
	return node.WildCardChild.Find(route, currentIndex+1, urlParams)
}
