package router

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
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
	Handler http.Handler
	Context RequestContext
}

type tree struct {
	nodes [5]treeNode
}

type routePart struct {
	route    string
	wildcard bool
}

func NewTree() *tree {
	return &tree{
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

	// Flag wildcard parameters and check potential duplication
	routeMembers := make([]routePart, len(routeSplit))
	wildcards := make(map[string]struct{})
	for i, item := range routeSplit {
		if item[0] != WILDCARD_START_CHAR {
			routeMembers[i] = routePart{item, false}
			continue
		}

		// Handle wildcard
		if _, found := wildcards[item]; found {
			panic(fmt.Sprintf("[%s] %s found duplicated wildcard parameter name: %s", method, route, item))
		}
		wildcards[item] = struct{}{}
		// Remove '{' & '}'
		routeMembers[i] = routePart{item[1 : len(item)-1], true}
	}

	err := root.Register(routeMembers, 0, handler)
	if err != nil {
		panic(fmt.Sprintf("[%s] %s %v", method, route, err))
	}
}

func (t *tree) Find(method HttpMethod, url *url.URL) (RouteData, error) {
	root, found := t.GetRootNode(method)
	if !found {
		return RouteData{}, ErrUnhandledMethod
	}

	routeSplit := strings.FieldsFunc(url.Path, splitFn)
	if len(routeSplit) == 0 {
		// Root path
		if root.Handler != nil {
			return RouteData{Handler: root.Handler}, nil
		} else {
			return RouteData{}, ErrNotFound
		}
	}

	routeParams := map[string]string{}
	node, found := root.Find(routeSplit, 0, routeParams)
	if !found {
		return RouteData{}, ErrNotFound
	}

	return RouteData{node.Handler, RequestContext{routeParams, url.Query()}}, nil
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
func (node *treeNode) Register(route []routePart, currentIndex int, handler http.Handler) error {
	var currentNode *treeNode
	var isWildcard bool
	if !route[currentIndex].wildcard {
		// Normal node
		currentNode = node.Children[route[currentIndex].route]
	} else {
		// Wildcard node
		for _, wilcardNode := range node.WildCardChildren {
			if wilcardNode.Content == route[currentIndex].route {
				currentNode = wilcardNode
				break
			}
		}
		isWildcard = true
	}

	if currentNode == nil {
		// New node
		currentNode = &treeNode{
			Content:          route[currentIndex].route,
			Children:         make(map[string]*treeNode),
			WildCardChildren: []*treeNode{},
		}

		if isWildcard {
			node.WildCardChildren = append(node.WildCardChildren, currentNode)
		} else {
			node.Children[route[currentIndex].route] = currentNode
		}

		if currentIndex == len(route)-1 {
			// Register handler on final node
			currentNode.Handler = handler
			return nil
		}
	} else if currentIndex == len(route)-1 {
		// Last node exists
		if currentNode.Handler == nil {
			// Register handler on final node
			currentNode.Handler = handler
			return nil
		}

		// Handler already registered on this node: panic
		return errors.New("route was already registered with another handler on the same HTTP method")
	}

	return currentNode.Register(route, currentIndex+1, handler)
}

func (node *treeNode) Find(route []string, currentIndex int, routeParams map[string]string) (*treeNode, bool) {
	if currentIndex == len(route) {
		// Last index: try find handler
		if node.Handler != nil {
			return node, true
		} else {
			return nil, false
		}
	}

	// Try find matching children
	if n, found := node.Children[route[currentIndex]]; found {
		foundNode, found := n.Find(route, currentIndex+1, routeParams) // Recursive find on matching node
		if found {
			return foundNode, true
		}
	}

	// No matching classic children: try wildcards
	for _, wildcardNode := range node.WildCardChildren {
		foundNode, found := wildcardNode.Find(route, currentIndex+1, routeParams) // Recursive find on wildcard node
		if found {
			// Populate url parameters
			routeParams[wildcardNode.Content] = route[currentIndex]
			return foundNode, true
		}
	}

	// Not found
	return nil, false
}
