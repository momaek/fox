// Copyright 2013 Julien Schmidt. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package fox

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// PerformRequest router test
func PerformRequest(r http.Handler, method, path string, header http.Header, body ...io.Reader) *httptest.ResponseRecorder {
	var data io.Reader
	if len(body) > 0 {
		data = body[0]
	}
	req := httptest.NewRequest(method, path, data)
	req.Header = header
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

type mockResponseWriter struct{}

func (m *mockResponseWriter) Header() (h http.Header) {
	return http.Header{}
}

func (m *mockResponseWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (m *mockResponseWriter) WriteString(s string) (n int, err error) {
	return len(s), nil
}

func (m *mockResponseWriter) WriteHeader(int) {}

func TestParams(t *testing.T) {
	ps := Params{
		Param{"param1", "value1"},
		Param{"param2", "value2"},
		Param{"param3", "value3"},
	}
	for i := range ps {
		if val := ps.ByName(ps[i].Key); val != ps[i].Value {
			t.Errorf("Wrong value for %s: Got %s; Want %s", ps[i].Key, val, ps[i].Value)
		}
	}
	if val := ps.ByName("noKey"); val != "" {
		t.Errorf("Expected empty string for not found key; got: %s", val)
	}
}

func TestRouter(t *testing.T) {
	router := New()

	routed := false
	router.Handle(http.MethodGet, "/user/:name", func(c *Context) {
		routed = true
		want := &Params{Param{"name", "gopher"}}
		if !reflect.DeepEqual(c.Params, want) {
			t.Fatalf("wrong wildcard values: want %v, got %v", want, c.Params)
		}
	})

	w := new(mockResponseWriter)

	req, _ := http.NewRequest(http.MethodGet, "/user/gopher", nil)
	router.ServeHTTP(w, req)

	if !routed {
		t.Fatal("routing failed")
	}
}

type handlerStruct struct {
	handled *bool
}

func (h handlerStruct) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	*h.handled = true
}

func TestRouterAPI(t *testing.T) {
	var get, head, options, post, put, patch, delete bool

	router := New()
	router.GET("/GET", func(c *Context) {
		get = true
	})
	router.HEAD("/GET", func(c *Context) {
		head = true
	})
	router.OPTIONS("/GET", func(c *Context) {
		options = true
	})
	router.POST("/POST", func(c *Context) {
		post = true
	})
	router.PUT("/PUT", func(c *Context) {
		put = true
	})
	router.PATCH("/PATCH", func(c *Context) {
		patch = true
	})
	router.DELETE("/DELETE", func(c *Context) {
		delete = true
	})

	w := new(mockResponseWriter)

	r, _ := http.NewRequest(http.MethodGet, "/GET", nil)
	router.ServeHTTP(w, r)
	if !get {
		t.Error("routing GET failed")
	}

	r, _ = http.NewRequest(http.MethodHead, "/GET", nil)
	router.ServeHTTP(w, r)
	if !head {
		t.Error("routing HEAD failed")
	}

	r, _ = http.NewRequest(http.MethodOptions, "/GET", nil)
	router.ServeHTTP(w, r)
	if !options {
		t.Error("routing OPTIONS failed")
	}

	r, _ = http.NewRequest(http.MethodPost, "/POST", nil)
	router.ServeHTTP(w, r)
	if !post {
		t.Error("routing POST failed")
	}

	r, _ = http.NewRequest(http.MethodPut, "/PUT", nil)
	router.ServeHTTP(w, r)
	if !put {
		t.Error("routing PUT failed")
	}

	r, _ = http.NewRequest(http.MethodPatch, "/PATCH", nil)
	router.ServeHTTP(w, r)
	if !patch {
		t.Error("routing PATCH failed")
	}

	r, _ = http.NewRequest(http.MethodDelete, "/DELETE", nil)
	router.ServeHTTP(w, r)
	if !delete {
		t.Error("routing DELETE failed")
	}
}

func TestRouterInvalidInput(t *testing.T) {
	router := New()

	router.basePath = ""

	handle := func(*Context) {}

	recv := catchPanic(func() {
		router.Handle("", "/", handle)
	})
	if recv == nil {
		t.Fatal("registering empty method did not panic")
	}

	recv = catchPanic(func() {
		router.GET("", handle)
	})
	if recv == nil {
		t.Fatal("registering empty path did not panic")
	}

	recv = catchPanic(func() {
		router.GET("noSlashRoot", handle)
	})
	if recv == nil {
		t.Fatal("registering path not beginning with '/' did not panic")
	}

	recv = catchPanic(func() {
		router.GET("/", nil)
	})
	if recv == nil {
		t.Fatal("registering nil handler did not panic")
	}
}

func TestRouterChaining(t *testing.T) {
	router1 := New()
	router2 := New()

	fooHit := false
	router1.POST("/foo", func(c *Context) {
		fooHit = true
		c.Writer.WriteHeader(http.StatusOK)
	})

	barHit := false
	router2.POST("/bar", func(c *Context) {
		barHit = true
		c.Writer.WriteHeader(http.StatusOK)
	})

	r, _ := http.NewRequest(http.MethodPost, "/foo", nil)
	w := httptest.NewRecorder()
	router1.ServeHTTP(w, r)
	if !(w.Code == http.StatusOK && fooHit) {
		t.Errorf("Regular routing failed with router chaining.")
		t.FailNow()
	}

	r, _ = http.NewRequest(http.MethodPost, "/bar", nil)
	w = httptest.NewRecorder()
	router1.ServeHTTP(w, r)
	if !(w.Code == http.StatusOK && barHit) {
		t.Errorf("Chained routing failed with router chaining.")
		t.FailNow()
	}

	r, _ = http.NewRequest(http.MethodPost, "/qax", nil)
	w = httptest.NewRecorder()
	router1.ServeHTTP(w, r)
	if !(w.Code == http.StatusNotFound) {
		t.Errorf("NotFound behavior failed with router chaining.")
		t.FailNow()
	}
}

func BenchmarkAllowed(b *testing.B) {
	handlerFunc := func(*Context) {}

	router := New()
	router.POST("/path", handlerFunc)
	router.GET("/path", handlerFunc)

	b.Run("Global", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = router.allowed("*", http.MethodOptions)
		}
	})
	b.Run("Path", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = router.allowed("/path", http.MethodOptions)
		}
	})
}

func TestRouterNotAllowed(t *testing.T) {
	handlerFunc := func(*Context) {}

	router := New()
	router.POST("/path", handlerFunc)

	// test not allowed
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if !(w.Code == http.StatusMethodNotAllowed) {
		t.Errorf("NotAllowed handling failed: Code=%d, Header=%v", w.Code, w.Header())
	} else if allow := w.Header().Get("Allow"); allow != "OPTIONS, POST" {
		t.Error("unexpected Allow header value: " + allow)
	}

	// add another method
	router.DELETE("/path", handlerFunc)
	router.OPTIONS("/path", handlerFunc) // must be ignored

	// test again
	r, _ = http.NewRequest(http.MethodGet, "/path", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if !(w.Code == http.StatusMethodNotAllowed) {
		t.Errorf("NotAllowed handling failed: Code=%d, Header=%v", w.Code, w.Header())
	} else if allow := w.Header().Get("Allow"); allow != "DELETE, OPTIONS, POST" {
		t.Error("unexpected Allow header value: " + allow)
	}

	// test custom handler
	w = httptest.NewRecorder()
	responseText := "custom method"

	router.NoMethod(func(c *Context) (string, int) {
		return responseText, http.StatusTeapot
	})

	router.ServeHTTP(w, r)
	if got := w.Body.String(); !(got == responseText) {
		t.Errorf("unexpected response got %q want %q", got, responseText)
	}
	if w.Code != http.StatusTeapot {
		t.Errorf("unexpected response code %d want %d", w.Code, http.StatusTeapot)
	}
	if allow := w.Header().Get("Allow"); allow != "DELETE, OPTIONS, POST" {
		t.Error("unexpected Allow header value: " + allow)
	}
}

func TestRouterNotFound(t *testing.T) {
	assert := assert.New(t)
	handlerFunc := func(*Context) {}

	router := New()
	router.RedirectFixedPath = true
	router.GET("/path", handlerFunc)
	router.GET("/dir/", handlerFunc)
	router.GET("/", handlerFunc)

	testRoutes := []struct {
		route    string
		code     int
		location string
	}{
		{"/path/", http.StatusMovedPermanently, "/path"},   // TSR -/
		{"/dir", http.StatusMovedPermanently, "/dir/"},     // TSR +/
		{"/PATH", http.StatusMovedPermanently, "/path"},    // Fixed Case
		{"/DIR/", http.StatusMovedPermanently, "/dir/"},    // Fixed Case
		{"/PATH/", http.StatusMovedPermanently, "/path"},   // Fixed Case -/
		{"/DIR", http.StatusMovedPermanently, "/dir/"},     // Fixed Case +/
		{"/../path", http.StatusMovedPermanently, "/path"}, // CleanPath
		{"/nope", http.StatusNotFound, ""},                 // NotFound
	}
	for _, tr := range testRoutes {
		w := PerformRequest(router, http.MethodGet, tr.route, nil)
		if !(w.Code == tr.code && (w.Code == http.StatusNotFound || fmt.Sprint(w.Header().Get("Location")) == tr.location)) {
			t.Errorf("NotFound handling route %s failed: Code=%d, Header=%v", tr.route, w.Code, w.Header().Get("Location"))
		}
	}

	// Test custom not found handler
	var notFound bool

	router.NotFound(func(c *Context) (interface{}, int) {
		notFound = true
		return nil, 404
	})

	w := PerformRequest(router, http.MethodGet, "/nope", nil)
	assert.Equal(http.StatusNotFound, w.Code)
	assert.True(notFound)

	// Test other method than GET (want 308 instead of 301)
	router.PATCH("/path", handlerFunc)
	w = PerformRequest(router, http.MethodPatch, "/path/", nil)
	// fmt.Printf("----w: %#v\n", w)

	assert.Equal(http.StatusPermanentRedirect, w.Code)
	// assert.Equal(t, "map[Location:[/path]]", fmt.Sprint(w.Header()))

	// Test special case where no node for the prefix "/" exists
	router = New()
	router.GET("/a", handlerFunc)
	w = PerformRequest(router, http.MethodGet, "/", nil)
	if !(w.Code == http.StatusNotFound) {
		t.Errorf("NotFound handling route / failed: Code=%d", w.Code)
	}
}

func TestRouterPanicHandler(t *testing.T) {
	router := New()
	panicHandled := false

	router.PanicHandler = func(rw http.ResponseWriter, r *http.Request, p interface{}) {
		panicHandled = true
	}

	router.Handle(http.MethodPut, "/user/:name", func(*Context) {
		panic("oops!")
	})

	w := new(mockResponseWriter)
	req, _ := http.NewRequest(http.MethodPut, "/user/gopher", nil)

	defer func() {
		if rcv := recover(); rcv != nil {
			t.Fatal("handling panic failed")
		}
	}()

	router.ServeHTTP(w, req)

	if !panicHandled {
		t.Fatal("simulating failed")
	}
}

type mockFileSystem struct {
	opened bool
}

func (mfs *mockFileSystem) Open(name string) (http.File, error) {
	mfs.opened = true
	return nil, errors.New("this is just a mock")
}

func TestRouterServeFiles(t *testing.T) {
	router := New()
	mfs := &mockFileSystem{}

	recv := catchPanic(func() {
		router.ServeFiles("/noFilepath", mfs)
	})
	if recv == nil {
		t.Fatal("registering path not ending with '*filepath' did not panic")
	}

	router.ServeFiles("/*filepath", mfs)
	w := new(mockResponseWriter)
	r, _ := http.NewRequest(http.MethodGet, "/favicon.ico", nil)
	router.ServeHTTP(w, r)
	if !mfs.opened {
		t.Error("serving file failed")
	}
}
