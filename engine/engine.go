package engine

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
)

// HandlerFunc is a function that can be registered to a route to handle HTTP
// requests. Like http.HandlerFunc, but has a third parameter for the values of
// wildcards (path variables).
type HandlerFunc func(http.ResponseWriter, *http.Request, Params)

// Make sure the Router conforms with the http.Handler interface
var _ http.Handler = New()

// Engine http server
type Engine struct {
	Router

	// If enabled, adds the matched route path onto the http.Request context
	// before invoking the handler.
	// The matched route path is only added to handlers of routes that were
	// registered when this option was enabled.
	SaveMatchedRoutePath bool

	// Enables automatic redirection if the current route can't be matched but a
	// handler for the path with (without) the trailing slash exists.
	// For example if /foo/ is requested but a route only exists for /foo, the
	// client is redirected to /foo with http status code 301 for GET requests
	// and 308 for all other request methods.
	RedirectTrailingSlash bool

	// If enabled, the router tries to fix the current request path, if no
	// handle is registered for it.
	// First superfluous path elements like ../ or // are removed.
	// Afterwards the router does a case-insensitive lookup of the cleaned path.
	// If a handle can be found for this route, the router makes a redirection
	// to the corrected path with status code 301 for GET requests and 308 for
	// all other request methods.
	// For example /FOO and /..//Foo could be redirected to /foo.
	// RedirectTrailingSlash is independent of this option.
	RedirectFixedPath bool

	// If enabled, the router checks if another method is allowed for the
	// current route, if the current request can not be routed.
	// If this is the case, the request is answered with 'Method Not Allowed'
	// and HTTP status code 405.
	// If no other Method is allowed, the request is delegated to the NotFound
	// handler.
	HandleMethodNotAllowed bool

	// If enabled, the router automatically replies to OPTIONS requests.
	// Custom OPTIONS handlers take priority over automatic replies.
	HandleOPTIONS bool

	// An optional http.Handler that is called on automatic OPTIONS requests.
	// The handler is only called if HandleOPTIONS is true and no OPTIONS
	// handler for the specific path was set.
	// The "Allowed" header is set before calling the handler.
	GlobalOPTIONS http.Handler

	// Cached value of global (*) allowed methods
	globalAllowed string

	// Configurable http.Handler which is called when no matching route is
	// found. If it is not set, http.NotFound is used.
	NotFound http.Handler

	// Configurable http.Handler which is called when a request
	// cannot be routed and HandleMethodNotAllowed is true.
	// If it is not set, http.Error with http.StatusMethodNotAllowed is used.
	// The "Allow" header with allowed request methods is set before the handler
	// is called.
	MethodNotAllowed http.Handler

	// Function to handle panics recovered from http handlers.
	// It should be used to generate a error page and return the http error code
	// 500 (Internal Server Error).
	// The handler can be used to keep your server from crashing because of
	// unrecovered panics.
	PanicHandler func(http.ResponseWriter, *http.Request, interface{})

	paramsPool sync.Pool
}

// New returns a new initialized Engine.
// Path auto-correction, including trailing slashes, is enabled by default.
func New() *Engine {

	engine := &Engine{
		RedirectTrailingSlash:  true,
		RedirectFixedPath:      true,
		HandleMethodNotAllowed: true,
		HandleOPTIONS:          true,
		Router:                 Router{},
	}

	engine.Router.engine = engine

	return engine
}

func (engine *Engine) getParams() *Params {
	ps, _ := engine.paramsPool.Get().(*Params)
	*ps = (*ps)[0:0] // reset slice
	return ps
}

func (engine *Engine) putParams(ps *Params) {
	if ps != nil {
		engine.paramsPool.Put(ps)
	}
}

// Run attaches the router to a http.Server and starts listening and serving HTTP requests.
// It is a shortcut for http.ListenAndServe(addr, router)
// Note: this method will block the calling goroutine indefinitely unless an error happens.
func (engine *Engine) Run(addr string) (err error) {

	err = http.ListenAndServe(addr, engine)
	return
}

// RunTLS attaches the router to a http.Server and starts listening and serving HTTPS (secure) requests.
// It is a shortcut for http.ListenAndServeTLS(addr, certFile, keyFile, router)
// Note: this method will block the calling goroutine indefinitely unless an error happens.
func (engine *Engine) RunTLS(addr, certFile, keyFile string) (err error) {

	err = http.ListenAndServeTLS(addr, certFile, keyFile, engine)
	return
}

// RunUnix attaches the router to a http.Server and starts listening and serving HTTP requests
// through the specified unix socket (ie. a file).
// Note: this method will block the calling goroutine indefinitely unless an error happens.
func (engine *Engine) RunUnix(file string) (err error) {

	listener, err := net.Listen("unix", file)
	if err != nil {
		return
	}
	defer listener.Close()
	defer os.Remove(file)

	err = http.Serve(listener, engine)
	return
}

// RunFd attaches the router to a http.Server and starts listening and serving HTTP requests
// through the specified file descriptor.
// Note: this method will block the calling goroutine indefinitely unless an error happens.
func (engine *Engine) RunFd(fd int) (err error) {

	f := os.NewFile(uintptr(fd), fmt.Sprintf("fd@%d", fd))
	listener, err := net.FileListener(f)
	if err != nil {
		return
	}
	defer listener.Close()
	err = engine.RunListener(listener)
	return
}

// RunListener attaches the router to a http.Server and starts listening and serving HTTP requests
// through the specified net.Listener
func (engine *Engine) RunListener(listener net.Listener) (err error) {

	err = http.Serve(listener, engine)
	return
}

// ServeHTTP makes the router implement the http.Handler interface.
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	engine.handleHTTPRequest(w, req)
}

func (engine *Engine) recv(w http.ResponseWriter, req *http.Request) {
	if rcv := recover(); rcv != nil {
		engine.PanicHandler(w, req, rcv)
	}
}

// handleHTTPRequest makes the router implement the http.Handler interface.
func (engine *Engine) handleHTTPRequest(w http.ResponseWriter, req *http.Request) {
	if engine.PanicHandler != nil {
		defer engine.recv(w, req)
	}

	path := req.URL.Path

	if root := engine.Router.trees[req.Method]; root != nil {
		if handle, ps, tsr := root.getValue(path, engine.getParams); handle != nil {
			if ps != nil {
				handle(w, req, *ps)
				engine.putParams(ps)
			} else {
				handle(w, req, nil)
			}
			return
		} else if req.Method != http.MethodConnect && path != "/" {
			// Moved Permanently, request with GET method
			code := http.StatusMovedPermanently
			if req.Method != http.MethodGet {
				// Permanent Redirect, request with same method
				code = http.StatusPermanentRedirect
			}

			if tsr && engine.RedirectTrailingSlash {
				if len(path) > 1 && path[len(path)-1] == '/' {
					req.URL.Path = path[:len(path)-1]
				} else {
					req.URL.Path = path + "/"
				}
				http.Redirect(w, req, req.URL.String(), code)
				return
			}

			// Try to fix the request path
			if engine.RedirectFixedPath {
				fixedPath, found := root.findCaseInsensitivePath(
					CleanPath(path),
					engine.RedirectTrailingSlash,
				)
				if found {
					req.URL.Path = fixedPath
					http.Redirect(w, req, req.URL.String(), code)
					return
				}
			}
		}
	}

	if req.Method == http.MethodOptions && engine.HandleOPTIONS {
		// Handle OPTIONS requests
		if allow := engine.Router.allowed(path, http.MethodOptions); allow != "" {
			w.Header().Set("Allow", allow)
			if engine.GlobalOPTIONS != nil {
				engine.GlobalOPTIONS.ServeHTTP(w, req)
			}
			return
		}
	} else if engine.HandleMethodNotAllowed { // Handle 405
		if allow := engine.Router.allowed(path, req.Method); allow != "" {
			w.Header().Set("Allow", allow)
			if engine.MethodNotAllowed != nil {
				engine.MethodNotAllowed.ServeHTTP(w, req)
			} else {
				http.Error(w,
					http.StatusText(http.StatusMethodNotAllowed),
					http.StatusMethodNotAllowed,
				)
			}
			return
		}
	}

	// Handle 404
	if engine.NotFound != nil {
		engine.NotFound.ServeHTTP(w, req)
	} else {
		http.NotFound(w, req)
	}
}
