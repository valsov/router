# HTTP router mux

Implementation of a request router/dispatcher based on a **radix tree** to enable fast route lookup.

## Usage

```go
func main() {
	// Configure router
	mux := router.NewHttpRouter()
	
	// Handle with http.HandlerFunc
	mux.HandleFunc(router.GET, "/route/to/handle1", func(w http.ResponseWriter, r *http.Request) {
		// [...]
	})
	
	// Handle with http.Handler
	handler := CustomHandler{}
	mux.Handle(router.GET, "/route/to/handle2", &handler)
	
	// Support wildcards in path
	mux.HandleFunc(router.GET, "/route/{param1}/sample/{param2}", func(w http.ResponseWriter, r *http.Request) {
		// [...]
	})
	
	// Add middleware
	mux.UseMiddleware(middleware.LoggerMiddleware(loggerInstance))
	err := http.ListenAndServe("addr", mux)
	// [...]
}

type CustomHandler struct{}

func (c *CustomHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// [...]
}
```

## Radix tree

Each registered route is split to form a tree, with the HTTP method as a route node. **Wildcards** are supported using the following syntax: `{wildcardName}`.
Example:
```mermaid

graph TD;

GET-->getRoot(/);

getRoot-->getUsers(users);

getUsers-->getUserWithId("users/{userId}");

getRoot-->stats;

stats-->sample;

POST-->postRoot(/);

postRoot-->postUsers(users);

```


