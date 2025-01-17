package fox

import (
	"net/http"
	"sync"
	"time"

	"github.com/miclle/fox/render"
)

// Context allows us to pass variables between middleware,
// manage the flow, using logger with context
type Context struct {
	Request *http.Request
	Writer  *ResponseWriter
	Params  *Params

	engine   *Engine
	handlers HandlersChain
	index    int

	// This mutex protects Keys map.
	mu sync.RWMutex

	// Keys is a key/value pair exclusively for the context of each request.
	Keys map[string]any
}

func (c *Context) reset(w http.ResponseWriter, req *http.Request) {
	c.Writer = &ResponseWriter{
		ResponseWriter: w,
		size:           noWritten,
		status:         defaultStatus,
	}
	c.Request = req
	*c.Params = (*c.Params)[:0]
	c.handlers = nil
	c.index = -1
	c.Keys = nil
}

// Next should be used only inside middleware.
func (c *Context) Next() {
	c.index++
	for c.index < len(c.handlers) {
		res, code, err := call(c, c.handlers[c.index])
		if err != nil {
			c.renderError(err)
			return
		}
		if res != nil || code != 0 {
			c.render(code, res)
		}
		c.index++
	}
}

// renderError ...
func (c *Context) renderError(err error) {
	c.Writer.WriteHeader(http.StatusInternalServerError)
	c.Writer.Write([]byte(err.Error()))
}

// render writes the response headers and calls render.render to render data.
func (c *Context) render(code int, res any) {

	if code > 0 {
		c.Writer.WriteHeader(code)
	}

	var r render.Render
	switch v := res.(type) {
	case error:
		c.renderError(v)
		return
	case string:
		r = render.String{Format: v}
	case render.Redirect:
		r = v
		c.Writer.WriteHeader(-1)
	case render.JSON, render.IndentedJSON, render.JsonpJSON, render.XML, render.Data,
		render.HTML, render.YAML, render.Reader, render.ASCIIJSON, render.ProtoBuf:
		r = v.(render.Render)
	default:
		r = render.JSON{Data: res}
	}

	r.WriteContentType(c.Writer)
	if err := r.Render(c.Writer); err != nil {
		panic(err)
	}
}

/************************************/
/******** METADATA MANAGEMENT********/
/************************************/

// TODO(m) Using Generics

// Set is used to store a new key/value pair exclusively for this context.
// It also lazy initializes  c.Keys if it was not used previously.
func (c *Context) Set(key string, value any) {
	c.mu.Lock()
	if c.Keys == nil {
		c.Keys = make(map[string]any)
	}

	c.Keys[key] = value
	c.mu.Unlock()
}

// Get returns the value for the given key, ie: (value, true).
// If the value does not exist it returns (nil, false)
func (c *Context) Get(key string) (value any, exists bool) {
	c.mu.RLock()
	value, exists = c.Keys[key]
	c.mu.RUnlock()
	return
}

// MustGet returns the value for the given key if it exists, otherwise it panics.
func (c *Context) MustGet(key string) any {
	if value, exists := c.Get(key); exists {
		return value
	}
	panic("Key \"" + key + "\" does not exist")
}

/************************************/
/**** HTTPS://PKG.GO.DEV/CONTEXT ****/
/************************************/

// Deadline returns that there is no deadline (ok==false) when c.Request has no Context.
func (c *Context) Deadline() (deadline time.Time, ok bool) {
	if c.Request == nil || c.Request.Context() == nil {
		return
	}
	return c.Request.Context().Deadline()
}

// Done returns nil (chan which will wait forever) when c.Request has no Context.
func (c *Context) Done() <-chan struct{} {
	if c.Request == nil || c.Request.Context() == nil {
		return nil
	}
	return c.Request.Context().Done()
}

// Err returns nil when c.Request has no Context.
func (c *Context) Err() error {
	if c.Request == nil || c.Request.Context() == nil {
		return nil
	}
	return c.Request.Context().Err()
}

// Value returns the value associated with this context for key, or nil
// if no value is associated with key. Successive calls to Value with
// the same key returns the same result.
func (c *Context) Value(key any) any {
	if key == 0 {
		return c.Request
	}
	if keyAsString, ok := key.(string); ok {
		if val, exists := c.Get(keyAsString); exists {
			return val
		}
	}
	if c.Request == nil || c.Request.Context() == nil {
		return nil
	}
	return c.Request.Context().Value(key)
}
