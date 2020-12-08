package engine

import (
	"net/http"
)

// Mode env
type Mode string

const (
	// DevelopmentMode indicates engine mode is development.
	DevelopmentMode Mode = "development"

	// ProductionMode indicates engine mode is production.
	ProductionMode Mode = "production"

	// TestMode indicates engine mode is test.
	TestMode Mode = "test"
)

var engineMode = DevelopmentMode

// SetMode sets gin mode according to input string.
func SetMode(value Mode) {
	switch value {
	case DevelopmentMode:
		engineMode = DevelopmentMode
	case ProductionMode:
		engineMode = ProductionMode
	case TestMode:
		engineMode = TestMode
	default:
		panic("engine mode unknown: " + value)
	}
}

// HandlerFunc defines the handler used by middleware as return value.
type HandlerFunc func(Context) (res interface{}, err error)

// HandlersChain defines a HandlerFunc array.
type HandlersChain []HandlerFunc

// Last returns the last handler in the chain. ie. the last handler is the main one.
func (c HandlersChain) Last() HandlerFunc {
	if length := len(c); length > 0 {
		return c[length-1]
	}
	return nil
}

// RouterConfigFunc engine load router config func
type RouterConfigFunc func(router IRouter)

// Engine http server
type Engine interface {
	IRouter

	// NoRoute adds handlers for NoRoute. It is recommended to return a 404 code by default
	NoRoute(handlers ...HandlerFunc)

	// Load router config
	Load(f RouterConfigFunc)

	// Run attaches the router to a http.Server and starts listening and serving HTTP requests.
	// It is a shortcut for http.ListenAndServe(addr, router)
	// Note: this method will block the calling goroutine indefinitely unless an error happens.
	Run(addr string) (err error)

	// RunTLS attaches the router to a http.Server and starts listening and serving HTTPS (secure) requests.
	// It is a shortcut for http.ListenAndServeTLS(addr, certFile, keyFile, router)
	// Note: this method will block the calling goroutine indefinitely unless an error happens.
	RunTLS(addr, certFile, keyFile string) (err error)

	// ServeHTTP conforms to the http.Handler interface.
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

// IRouter defines all router handle interface includes single and group router.
type IRouter interface {

	// Use middleware
	Use(middleware ...HandlerFunc)

	Handle(httpMethod string, relativePath string, handlers ...HandlerFunc)
	Any(relativePath string, handlers ...HandlerFunc)
	GET(relativePath string, handlers ...HandlerFunc)
	POST(relativePath string, handlers ...HandlerFunc)
	DELETE(relativePath string, handlers ...HandlerFunc)
	PATCH(relativePath string, handlers ...HandlerFunc)
	PUT(relativePath string, handlers ...HandlerFunc)
	OPTIONS(relativePath string, handlers ...HandlerFunc)
	HEAD(relativePath string, handlers ...HandlerFunc)

	Static(relativePath string, root string)
	StaticFile(relativePath string, filepath string)
	StaticFS(relativePath string, fs http.FileSystem)

	Group(relativePath string, handlers ...HandlerFunc) *IRouter
}
