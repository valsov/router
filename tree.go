package router

import (
	"errors"
	"fmt"
	"net/http"
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

type RouteData struct {
	Handler   http.Handler
	UrlParams map[string]string
	// todo: handle query params
}

type tree struct {
	nodes [5]treeNode
}

func New() tree {
	return tree{
		[5]treeNode{
			{Content: string(GET), Children: make(map[string]*treeNode), WildCardChildren: []*treeNode{}},
			{Content: string(POST), Children: make(map[string]*treeNode), WildCardChildren: []*treeNode{}},
			{Content: string(PUT), Children: make(map[string]*treeNode), WildCardChildren: []*treeNode{}},
			{Content: string(PATCH), Children: make(map[string]*treeNode), WildCardChildren: []*treeNode{}},
			{Content: string(DELETE), Children: make(map[string]*treeNode), WildCardChildren: []*treeNode{}},
		},
	}
}

// Can panic
func (t *tree) Register(method HttpMethod, route string, handler http.Handler) {
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

	// Check wildcard parameters duplication
	wildcards := make(map[string]struct{})
	for _, item := range routeSplit {
		if item[0] != WILDCARD_START_CHAR {
			continue
		}
		if _, found := wildcards[item]; found {
			panic(fmt.Sprintf("[%s] %s found duplicated wildcard parameter name: %s", method, route, item))
		}
		wildcards[item] = struct{}{}
	}

	root.Register(routeSplit, 0, handler)
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

type treeNode struct {
	Content          string
	Handler          http.Handler
	Children         map[string]*treeNode
	WildCardChildren []*treeNode
}

// Can panic
func (node *treeNode) Register(route []string, currentIndex int, handler http.Handler) {
	var currentNode *treeNode
	var isWildcard bool
	if route[currentIndex][0] != WILDCARD_START_CHAR {
		// Normal node
		currentNode = node.Children[route[currentIndex]]
	} else {
		// Wildcard node
		for _, wilcardNode := range node.WildCardChildren {
			if wilcardNode.Content == route[currentIndex] {
				currentNode = wilcardNode
				break
			}
		}
		isWildcard = true
	}

	if currentNode == nil {
		// New node
		currentNode = &treeNode{
			Content:          route[currentIndex],
			Children:         make(map[string]*treeNode),
			WildCardChildren: []*treeNode{},
		}

		if isWildcard {
			node.WildCardChildren = append(node.WildCardChildren, currentNode)
		} else {
			node.Children[route[currentIndex]] = currentNode
		}

		if currentIndex == len(route)-1 {
			// Register handler on final node
			node.Handler = handler
			return
		}
	} else if currentIndex == len(route)-1 {
		// Last node exists
		if currentNode.Handler == nil {
			// Register handler on final node
			node.Handler = handler
			return
		}

		// Handler already registered on this node: panic
		panic(fmt.Sprintf("%s was already registered with another handler on the same HTTP method", route))
	}

	currentNode.Register(route, currentIndex+1, handler)
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
