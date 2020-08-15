package engine

import (
	"fmt"
	"reflect"
)

// Group struct
type Group struct {
	app    *Engine
	prefix string
}

// Use registers a middleware route.
// Middleware matches requests beginning with the provided prefix.
// Providing a prefix is optional, it defaults to "/".
//
// - group.Use(handler)
// - group.Use("/api", handler)
// - group.Use("/api", handler, handler)
func (group *Group) Use(args ...interface{}) Router {
	var path = ""
	var handlers []Handler
	for i := 0; i < len(args); i++ {
		switch arg := args[i].(type) {
		case string:
			path = arg
		case Handler:
			handlers = append(handlers, arg)
		default:
			panic(fmt.Sprintf("use: invalid handler %v\n", reflect.TypeOf(arg)))
		}
	}
	group.app.register(methodUse, getGroupPath(group.prefix, path), handlers...)
	return group
}

// GET registers a route for GET methods that requests a representation
// of the specified resource. Requests using GET should only retrieve data.
func (group *Group) GET(path string, handlers ...Handler) Router {
	route := group.app.register("GET", getGroupPath(group.prefix, path), handlers...)
	// Add head route
	headRoute := route
	group.app.addRoute("HEAD", &headRoute)
	return group
}

// HEAD registers a route for HEAD methods that asks for a response identical
// to that of a GET request, but without the response body.
func (group *Group) HEAD(path string, handlers ...Handler) Router {
	return group.Add("HEAD", path, handlers...)
}

// POST registers a route for POST methods that is used to submit an entity to the
// specified resource, often causing a change in state or side effects on the server.
func (group *Group) POST(path string, handlers ...Handler) Router {
	return group.Add("POST", path, handlers...)
}

// PUT registers a route for PUT methods that replaces all current representations
// of the target resource with the request payload.
func (group *Group) PUT(path string, handlers ...Handler) Router {
	return group.Add("PUT", path, handlers...)
}

// DELETE registers a route for DELETE methods that deletes the specified resource.
func (group *Group) DELETE(path string, handlers ...Handler) Router {
	return group.Add("DELETE", path, handlers...)
}

// CONNECT registers a route for CONNECT methods that establishes a tunnel to the
// server identified by the target resource.
func (group *Group) CONNECT(path string, handlers ...Handler) Router {
	return group.Add("CONNECT", path, handlers...)
}

// OPTIONS registers a route for OPTIONS methods that is used to describe the
// communication options for the target resource.
func (group *Group) OPTIONS(path string, handlers ...Handler) Router {
	return group.Add("OPTIONS", path, handlers...)
}

// TRACE registers a route for TRACE methods that performs a message loop-back
// test along the path to the target resource.
func (group *Group) TRACE(path string, handlers ...Handler) Router {
	return group.Add("TRACE", path, handlers...)
}

// PATCH registers a route for PATCH methods that is used to apply partial
// modifications to a resource.
func (group *Group) PATCH(path string, handlers ...Handler) Router {
	return group.Add("PATCH", path, handlers...)
}

// Add ...
func (group *Group) Add(method, path string, handlers ...Handler) Router {
	group.app.register(method, getGroupPath(group.prefix, path), handlers...)
	return group
}

// Static ...
func (group *Group) Static(prefix, root string, config ...Static) Router {
	group.app.registerStatic(getGroupPath(group.prefix, prefix), root, config...)
	return group
}

// All ...
func (group *Group) All(path string, handlers ...Handler) Router {
	for _, method := range intMethod {
		group.Add(method, path, handlers...)
	}
	return group
}

// Group is used for Routes with common prefix to define a new sub-router with optional middleware.
func (group *Group) Group(prefix string, handlers ...Handler) Router {
	prefix = getGroupPath(group.prefix, prefix)
	if len(handlers) > 0 {
		group.app.register(methodUse, prefix, handlers...)
	}
	return group.app.Group(prefix)
}
