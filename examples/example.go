package main

import (
	"fmt"

	"github.com/valyala/fasthttp"
)

// MyHandler type
type MyHandler struct {
	foobar string
}

// HandleFastHTTP request handler in net/http style, i.e. method bound to MyHandler struct.
func (h *MyHandler) HandleFastHTTP(ctx *fasthttp.RequestCtx) {
	// notice that we may access MyHandler properties here - see h.foobar.
	fmt.Fprintf(ctx, "Hello, world! Requested path is %q. Foobar is %q",
		ctx.Path(), h.foobar)
}

// request handler in fasthttp style, i.e. just plain function.
func fastHTTPHandler(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, "Hi there! RequestURI is %q", ctx.RequestURI())
}

func main() {
	// pass bound struct method to fasthttp
	myHandler := &MyHandler{
		foobar: "foobar",
	}
	fasthttp.ListenAndServe(":8080", myHandler.HandleFastHTTP)

	// pass plain function to fasthttp
	fasthttp.ListenAndServe(":8081", fastHTTPHandler)

}
