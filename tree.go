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

	WILDCARD_START_CHAR byte = '{'
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
	// todo: handle query params
}

type tree struct {
	nodes [5]treeNode
}

func New() tree {
	return tree{
		[5]treeNode{
			{Content: string(GET), Children: map[string]*treeNode{}, WildCardChildren: []*treeNode{}},
			{Content: string(POST), Children: map[string]*treeNode{}, WildCardChildren: []*treeNode{}},
			{Content: string(PUT), Children: map[string]*treeNode{}, WildCardChildren: []*treeNode{}},
			{Content: string(PATCH), Children: map[string]*treeNode{}, WildCardChildren: []*treeNode{}},
			{Content: string(DELETE), Children: map[string]*treeNode{}, WildCardChildren: []*treeNode{}},
		},
	}
}

// Can panic
func (t *tree) Add(method HttpMethod, route string, handler RouteHandler) {
	root, found := t.getRootNode(method)
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
	root, found := t.getRootNode(method)
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

func (t *tree) getRootNode(method HttpMethod) (*treeNode, bool) {
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

type treeNode struct {
	Content          string
	Handler          RouteHandler
	Children         map[string]*treeNode
	WildCardChildren []*treeNode
}

// Can panic
func (node *treeNode) Add(route []string, currentIndex int, handler RouteHandler) {
	// todo: panic on duplicated urlparams

	//wildCard := route[currentIndex][0] == WILDCARD_START_CHAR
	//if wildCard

}

func (node *treeNode) Find(route []string, currentIndex int, urlParams map[string]string) (*treeNode, bool) {
	if currentIndex == len(route)-1 {
		// Last index: try find handler
		if node.Handler != nil {
			return node, true
		} else {
			return nil, false
		}
	}

	// Try find matching children
	if n, found := node.Children[route[currentIndex]]; found {
		foundNode, found := n.Find(route, currentIndex+1, urlParams) // Recursive find on matching node
		if found {
			return foundNode, true
		}
	}

	// No matching classic children: try wildcards
	for _, wildcardNode := range node.WildCardChildren {
		foundNode, found := wildcardNode.Find(route, currentIndex+1, urlParams) // Recursive find on wildcard node
		if found {
			// Populate url parameters
			urlParams[wildcardNode.Content] = route[currentIndex]
			return foundNode, true
		}
	}

	// Not found
	return nil, false
}
