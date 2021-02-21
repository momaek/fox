package engine

import (
	"log"
	"net/http"
	"path"
	"strings"
	"sync"
)

// cleanPath returns the canonical path for p, eliminating . and .. elements.
func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	np := path.Clean(p)
	// path.Clean removes trailing slash except for root;
	// put the trailing slash back if necessary.
	if p[len(p)-1] == '/' && np != "/" {
		// Fast path for common case of p being the string we want:
		if len(p) == len(np)+1 && strings.HasPrefix(p, np) {
			np = p
		} else {
			np += "/"
		}
	}
	return np
}

var _HTTPMethods = []string{
	http.MethodGet,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
	http.MethodHead,
	http.MethodOptions,
}

var _HTTPMethodMap = map[string]bool{
	http.MethodGet:     true,
	http.MethodPost:    true,
	http.MethodPut:     true,
	http.MethodPatch:   true,
	http.MethodDelete:  true,
	http.MethodHead:    true,
	http.MethodOptions: true,
}

type muxEntry struct {
	handlers []Handler
	pattern  string
	method   string
}

// Router is a http.Handler which can be used to dispatch requests to different
// handler functions via configurable routes
type Router struct {
	mu     sync.RWMutex
	es     []*muxEntry
	routes map[string]map[string]*muxEntry

	basePath string
}

func (router *Router) registered(httpMethod, pattern string) bool {
	if _, exist := router.routes[httpMethod][pattern]; exist {
		return true
	}

	if router.routes == nil {
		router.routes = make(map[string]map[string]*muxEntry)
	}

	if router.routes[httpMethod] == nil {
		router.routes[httpMethod] = make(map[string]*muxEntry)
	}

	// TODO: traverse the match

	return false
}

// Handle registers a new request handle with the given pattern, method and handlers.
func (router *Router) Handle(httpMethod, relativePath string, handlers ...Handler) {
	router.mu.Lock()
	defer router.mu.Unlock()

	httpMethod = strings.ToUpper(httpMethod)
	if !_HTTPMethodMap[httpMethod] {
		log.Panicf("unknown HTTP method: %s %s", httpMethod, relativePath)
	}

	relativePath = path.Join(router.basePath, relativePath)

	pattern := cleanPath(relativePath)
	// if pattern == "" {
	// 	panic("invalid pattern")
	// }

	if len(handlers) == 0 {
		log.Panicf("nil handler: %s %s", httpMethod, relativePath)
	}

	if exist := router.registered(httpMethod, pattern); exist {
		log.Panicf("multiple registrations: %s %s", httpMethod, relativePath)
	}

	entry := &muxEntry{
		method:   httpMethod,
		pattern:  pattern,
		handlers: handlers,
	}

	router.routes[httpMethod][pattern] = entry

	router.es = append(router.es, entry)
}

// Any registers a route that matches all the HTTP methods.
// GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS.
func (router *Router) Any(relativePath string, handlers ...Handler) {
	for _, method := range _HTTPMethods {
		router.Handle(method, relativePath, handlers...)
	}
}

// GET is a shortcut for router.Handle("GET", path, handle).
func (router *Router) GET(relativePath string, handlers ...Handler) {
	router.Handle(http.MethodGet, relativePath, handlers...)
}

// POST is a shortcut for router.Handle("POST", path, handle).
func (router *Router) POST(relativePath string, handlers ...Handler) {
	router.Handle(http.MethodPost, relativePath, handlers...)
}

// PUT is a shortcut for router.Handle("PUT", path, handle).
func (router *Router) PUT(relativePath string, handlers ...Handler) {
	router.Handle(http.MethodPut, relativePath, handlers...)
}

// PATCH is a shortcut for router.Handle("PATCH", path, handle).
func (router *Router) PATCH(relativePath string, handlers ...Handler) {
	router.Handle(http.MethodPatch, relativePath, handlers...)
}

// DELETE is a shortcut for router.Handle("DELETE", path, handle).
func (router *Router) DELETE(relativePath string, handlers ...Handler) {
	router.Handle(http.MethodDelete, relativePath, handlers...)
}

// HEAD is a shortcut for router.Handle("HEAD", path, handle).
func (router *Router) HEAD(relativePath string, handlers ...Handler) {
	router.Handle(http.MethodHead, relativePath, handlers...)
}

// OPTIONS is a shortcut for router.Handle("OPTIONS", path, handle).
func (router *Router) OPTIONS(relativePath string, handlers ...Handler) {
	router.Handle(http.MethodOptions, relativePath, handlers...)
}

// Group creates a new router group.
// You should add all the routes that have common middlewares or the same path prefix.
func (router *Router) Group(relativePath string, group func(group *Router)) {

	var r = &Router{
		// TODO(m)
		basePath: relativePath,
	}

	group(r)
}
