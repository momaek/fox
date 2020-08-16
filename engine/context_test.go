package engine

// go test -v -run=^$ -bench=Benchmark_Ctx_Accepts -benchmem -count=4
// go test -run Test_Ctx

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"text/template"
	"time"

	"github.com/valyala/bytebufferpool"
	"github.com/valyala/fasthttp"

	"fox/engine/utils"
)

// go test -run TestContextAccepts
func TestContextAccepts(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)
	ctx.Fasthttp.Request.Header.Set(HeaderAccept, "text/html,application/xhtml+xml,application/xml;q=0.9")
	utils.AssertEqual(t, "", ctx.Accepts(""))
	utils.AssertEqual(t, "", ctx.Accepts())
	utils.AssertEqual(t, ".xml", ctx.Accepts(".xml"))
	utils.AssertEqual(t, "", ctx.Accepts(".john"))
}

// go test -v -run=^$ -bench=Benchmark_Ctx_Accepts -benchmem -count=4
func Benchmark_Ctx_Accepts(b *testing.B) {
	app := New()
	c := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(c)
	c.Fasthttp.Request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9")
	var res string
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		res = c.Accepts(".xml")
	}
	utils.AssertEqual(b, ".xml", res)
}

// go test -run TestContextAccepts_EmptyAccept
func TestContextAccepts_EmptyAccept(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)
	utils.AssertEqual(t, ".forwarded", ctx.Accepts(".forwarded"))
}

// go test -run TestContextAccepts_Wildcard
func TestContextAccepts_Wildcard(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)
	ctx.Fasthttp.Request.Header.Set(HeaderAccept, "*/*;q=0.9")
	utils.AssertEqual(t, "html", ctx.Accepts("html"))
	utils.AssertEqual(t, "foo", ctx.Accepts("foo"))
	utils.AssertEqual(t, ".bar", ctx.Accepts(".bar"))
	ctx.Fasthttp.Request.Header.Set(HeaderAccept, "text/html,application/*;q=0.9")
	utils.AssertEqual(t, "xml", ctx.Accepts("xml"))
}

// go test -run TestContextAcceptsCharsets
func TestContextAcceptsCharsets(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)
	ctx.Fasthttp.Request.Header.Set(HeaderAcceptCharset, "utf-8, iso-8859-1;q=0.5")
	utils.AssertEqual(t, "utf-8", ctx.AcceptsCharsets("utf-8"))
}

// go test -v -run=^$ -bench=Benchmark_Ctx_AcceptsCharsets -benchmem -count=4
func Benchmark_Ctx_AcceptsCharsets(b *testing.B) {
	app := New()
	c := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(c)
	c.Fasthttp.Request.Header.Set("Accept-Charset", "utf-8, iso-8859-1;q=0.5")
	var res string
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		res = c.AcceptsCharsets("utf-8")
	}
	utils.AssertEqual(b, "utf-8", res)
}

// go test -run TestContextAcceptsEncodings
func TestContextAcceptsEncodings(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)
	ctx.Fasthttp.Request.Header.Set(HeaderAcceptEncoding, "deflate, gzip;q=1.0, *;q=0.5")
	utils.AssertEqual(t, "gzip", ctx.AcceptsEncodings("gzip"))
	utils.AssertEqual(t, "abc", ctx.AcceptsEncodings("abc"))
}

// go test -v -run=^$ -bench=Benchmark_Ctx_AcceptsEncodings -benchmem -count=4
func Benchmark_Ctx_AcceptsEncodings(b *testing.B) {
	app := New()
	c := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(c)
	c.Fasthttp.Request.Header.Set(HeaderAcceptEncoding, "deflate, gzip;q=1.0, *;q=0.5")
	var res string
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		res = c.AcceptsEncodings("gzip")
	}
	utils.AssertEqual(b, "gzip", res)
}

// go test -run TestContextAcceptsLanguages
func TestContextAcceptsLanguages(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)
	ctx.Fasthttp.Request.Header.Set(HeaderAcceptLanguage, "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5")
	utils.AssertEqual(t, "fr", ctx.AcceptsLanguages("fr"))
}

// go test -v -run=^$ -bench=Benchmark_Ctx_AcceptsLanguages -benchmem -count=4
func Benchmark_Ctx_AcceptsLanguages(b *testing.B) {
	app := New()
	c := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(c)
	c.Fasthttp.Request.Header.Set(HeaderAcceptLanguage, "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5")
	var res string
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		res = c.AcceptsLanguages("fr")
	}
	utils.AssertEqual(b, "fr", res)
}

// go test -run TestContextApp
func TestContextApp(t *testing.T) {
	t.Parallel()
	app := New()
	app.Settings.BodyLimit = 1000
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)
	utils.AssertEqual(t, 1000, ctx.App().Settings.BodyLimit)
}

// go test -run TestContextAppend
func TestContextAppend(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)
	ctx.Append("X-Test", "Hello")
	ctx.Append("X-Test", "World")
	ctx.Append("X-Test", "Hello", "World")
	// similar value in the middle
	ctx.Append("X2-Test", "World")
	ctx.Append("X2-Test", "XHello")
	ctx.Append("X2-Test", "Hello", "World")
	// similar value at the start
	ctx.Append("X3-Test", "XHello")
	ctx.Append("X3-Test", "World")
	ctx.Append("X3-Test", "Hello", "World")
	// try it with multiple similar values
	ctx.Append("X4-Test", "XHello")
	ctx.Append("X4-Test", "Hello")
	ctx.Append("X4-Test", "HelloZ")
	ctx.Append("X4-Test", "YHello")
	ctx.Append("X4-Test", "Hello")
	ctx.Append("X4-Test", "YHello")
	ctx.Append("X4-Test", "HelloZ")
	ctx.Append("X4-Test", "XHello")
	// without append value
	ctx.Append("X-Custom-Header")

	utils.AssertEqual(t, "Hello, World", string(ctx.Fasthttp.Response.Header.Peek("X-Test")))
	utils.AssertEqual(t, "World, XHello, Hello", string(ctx.Fasthttp.Response.Header.Peek("X2-Test")))
	utils.AssertEqual(t, "XHello, World, Hello", string(ctx.Fasthttp.Response.Header.Peek("X3-Test")))
	utils.AssertEqual(t, "XHello, Hello, HelloZ, YHello", string(ctx.Fasthttp.Response.Header.Peek("X4-Test")))
	utils.AssertEqual(t, "", string(ctx.Fasthttp.Response.Header.Peek("x-custom-header")))
}

// go test -v -run=^$ -bench=Benchmark_Ctx_Append -benchmem -count=4
func Benchmark_Ctx_Append(b *testing.B) {
	app := New()
	c := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(c)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		c.Append("X-Custom-Header", "Hello")
		c.Append("X-Custom-Header", "World")
		c.Append("X-Custom-Header", "Hello")
	}
	utils.AssertEqual(b, "Hello, World", getString(c.Fasthttp.Response.Header.Peek("X-Custom-Header")))
}

// go test -run TestContextAttachment
func TestContextAttachment(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)
	// empty
	ctx.Attachment()
	utils.AssertEqual(t, `attachment`, string(ctx.Fasthttp.Response.Header.Peek(HeaderContentDisposition)))
	// real filename
	ctx.Attachment("./static/img/logo.png")
	utils.AssertEqual(t, `attachment; filename="logo.png"`, string(ctx.Fasthttp.Response.Header.Peek(HeaderContentDisposition)))
	utils.AssertEqual(t, "image/png", string(ctx.Fasthttp.Response.Header.Peek(HeaderContentType)))
	// check quoting
	ctx.Attachment("another document.pdf\"\r\nBla: \"fasel")
	utils.AssertEqual(t, `attachment; filename="another+document.pdf%22%0D%0ABla%3A+%22fasel"`, string(ctx.Fasthttp.Response.Header.Peek(HeaderContentDisposition)))
}

// go test -v -run=^$ -bench=Benchmark_Ctx_Attachment -benchmem -count=4
func Benchmark_Ctx_Attachment(b *testing.B) {
	app := New()
	c := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(c)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		// example with quote params
		c.Attachment("another document.pdf\"\r\nBla: \"fasel")
	}
	utils.AssertEqual(b, `attachment; filename="another+document.pdf%22%0D%0ABla%3A+%22fasel"`, string(c.Fasthttp.Response.Header.Peek(HeaderContentDisposition)))
}

// go test -run TestContextBaseURL
func TestContextBaseURL(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)
	ctx.Fasthttp.Request.SetRequestURI("http://google.com/test")
	utils.AssertEqual(t, "http://google.com", ctx.BaseURL())
}

// go test -v -run=^$ -bench=Benchmark_Ctx_Append -benchmem -count=4
func Benchmark_Ctx_BaseURL(b *testing.B) {
	app := New()
	c := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(c)
	c.Fasthttp.Request.SetHost("google.com:1337")
	c.Fasthttp.Request.URI().SetPath("/haha/oke/lol")
	var res string
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		res = c.BaseURL()
	}
	utils.AssertEqual(b, "http://google.com:1337", res)
}

// go test -run TestContextBody
func TestContextBody(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)
	ctx.Fasthttp.Request.SetBody([]byte("john=doe"))
	utils.AssertEqual(t, "john=doe", ctx.Body())
}

// go test -run TestContextBodyParser
func TestContextBodyParser(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)

	type Demo struct {
		Name string `json:"name" xml:"name" form:"name" query:"name"`
	}

	testDecodeParser := func(contentType, body string) {
		ctx.Fasthttp.Request.Header.SetContentType(contentType)
		ctx.Fasthttp.Request.SetBody([]byte(body))
		ctx.Fasthttp.Request.Header.SetContentLength(len(body))
		d := new(Demo)
		utils.AssertEqual(t, nil, ctx.BodyParser(d))
		utils.AssertEqual(t, "john", d.Name)
	}

	testDecodeParser(MIMEApplicationJSON, `{"name":"john"}`)
	testDecodeParser(MIMEApplicationXML, `<Demo><name>john</name></Demo>`)
	testDecodeParser(MIMEApplicationJSON, `{"name":"john"}`)
	testDecodeParser(MIMEApplicationForm, "name=john")
	testDecodeParser(MIMEMultipartForm+`;boundary="b"`, "--b\r\nContent-Disposition: form-data; name=\"name\"\r\n\r\njohn\r\n--b--")

	testDecodeParserError := func(contentType, body string) {
		ctx.Fasthttp.Request.Header.SetContentType(contentType)
		ctx.Fasthttp.Request.SetBody([]byte(body))
		ctx.Fasthttp.Request.Header.SetContentLength(len(body))
		utils.AssertEqual(t, false, ctx.BodyParser(nil) == nil)
	}

	testDecodeParserError("invalid-content-type", "")
	testDecodeParserError(MIMEMultipartForm+`;boundary="b"`, "--b")
}

// go test -v -run=^$ -bench=Benchmark_Ctx_BodyParser_JSON -benchmem -count=4
func Benchmark_Ctx_BodyParser_JSON(b *testing.B) {
	app := New()
	c := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(c)
	type Demo struct {
		Name string `json:"name"`
	}
	body := []byte(`{"name":"john"}`)
	c.Fasthttp.Request.SetBody(body)
	c.Fasthttp.Request.Header.SetContentType(MIMEApplicationJSON)
	c.Fasthttp.Request.Header.SetContentLength(len(body))
	d := new(Demo)

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_ = c.BodyParser(d)
	}
	utils.AssertEqual(b, nil, c.BodyParser(d))
	utils.AssertEqual(b, "john", d.Name)
}

// go test -v -run=^$ -bench=Benchmark_Ctx_BodyParser_XML -benchmem -count=4
func Benchmark_Ctx_BodyParser_XML(b *testing.B) {
	app := New()
	c := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(c)
	type Demo struct {
		Name string `xml:"name"`
	}
	body := []byte("<Demo><name>john</name></Demo>")
	c.Fasthttp.Request.SetBody(body)
	c.Fasthttp.Request.Header.SetContentType(MIMEApplicationXML)
	c.Fasthttp.Request.Header.SetContentLength(len(body))
	d := new(Demo)

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_ = c.BodyParser(d)
	}
	utils.AssertEqual(b, nil, c.BodyParser(d))
	utils.AssertEqual(b, "john", d.Name)
}

// go test -v -run=^$ -bench=Benchmark_Ctx_BodyParser_Form -benchmem -count=4
func Benchmark_Ctx_BodyParser_Form(b *testing.B) {
	app := New()
	c := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(c)
	type Demo struct {
		Name string `form:"name"`
	}
	body := []byte("name=john")
	c.Fasthttp.Request.SetBody(body)
	c.Fasthttp.Request.Header.SetContentType(MIMEApplicationForm)
	c.Fasthttp.Request.Header.SetContentLength(len(body))
	d := new(Demo)

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_ = c.BodyParser(d)
	}
	utils.AssertEqual(b, nil, c.BodyParser(d))
	utils.AssertEqual(b, "john", d.Name)
}

// go test -v -run=^$ -bench=Benchmark_Ctx_BodyParser_MultipartForm -benchmem -count=4
func Benchmark_Ctx_BodyParser_MultipartForm(b *testing.B) {
	app := New()
	c := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(c)
	type Demo struct {
		Name string `form:"name"`
	}

	body := []byte("--b\r\nContent-Disposition: form-data; name=\"name\"\r\n\r\njohn\r\n--b--")
	c.Fasthttp.Request.SetBody(body)
	c.Fasthttp.Request.Header.SetContentType(MIMEMultipartForm + `;boundary="b"`)
	c.Fasthttp.Request.Header.SetContentLength(len(body))
	d := new(Demo)

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_ = c.BodyParser(d)
	}
	utils.AssertEqual(b, nil, c.BodyParser(d))
	utils.AssertEqual(b, "john", d.Name)
}

// go test -run TestContextContext
func TestContextContext(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)

	utils.AssertEqual(t, "*fasthttp.RequestCtx", fmt.Sprintf("%T", ctx.Context()))
}

// go test -run TestContextCookie
func TestContextCookie(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)
	expire := time.Now().Add(24 * time.Hour)
	var dst []byte
	dst = expire.In(time.UTC).AppendFormat(dst, time.RFC1123)
	httpdate := strings.Replace(string(dst), "UTC", "GMT", -1)
	ctx.Cookie(&Cookie{
		Name:    "username",
		Value:   "john",
		Expires: expire,
	})
	expect := "username=john; expires=" + httpdate + "; path=/; SameSite=Lax"
	utils.AssertEqual(t, expect, string(ctx.Fasthttp.Response.Header.Peek(HeaderSetCookie)))

	ctx.Cookie(&Cookie{SameSite: "strict"})
	ctx.Cookie(&Cookie{SameSite: "none"})
}

// go test -v -run=^$ -bench=Benchmark_Ctx_Cookie -benchmem -count=4
func Benchmark_Ctx_Cookie(b *testing.B) {
	app := New()
	c := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(c)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		c.Cookie(&Cookie{
			Name:  "John",
			Value: "Doe",
		})
	}
	utils.AssertEqual(b, "John=Doe; path=/; SameSite=Lax", getString(c.Fasthttp.Response.Header.Peek("Set-Cookie")))
}

// go test -run TestContextCookies
func TestContextCookies(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)
	ctx.Fasthttp.Request.Header.Set("Cookie", "john=doe")
	utils.AssertEqual(t, "doe", ctx.Cookies("john"))
	utils.AssertEqual(t, "default", ctx.Cookies("unknown", "default"))
}

// go test -run TestContextFormat
func TestContextFormat(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)
	ctx.Fasthttp.Request.Header.Set(HeaderAccept, MIMETextPlain)
	ctx.Format([]byte("Hello, World!"))
	utils.AssertEqual(t, "Hello, World!", string(ctx.Fasthttp.Response.Body()))

	ctx.Fasthttp.Request.Header.Set(HeaderAccept, MIMETextHTML)
	ctx.Format("Hello, World!")
	utils.AssertEqual(t, "<p>Hello, World!</p>", string(ctx.Fasthttp.Response.Body()))

	ctx.Fasthttp.Request.Header.Set(HeaderAccept, MIMEApplicationJSON)
	ctx.Format("Hello, World!")
	utils.AssertEqual(t, `"Hello, World!"`, string(ctx.Fasthttp.Response.Body()))
	ctx.Format(complex(1, 1))
	utils.AssertEqual(t, "(1+1i)", string(ctx.Fasthttp.Response.Body()))

	ctx.Fasthttp.Request.Header.Set(HeaderAccept, MIMEApplicationXML)
	ctx.Format("Hello, World!")
	utils.AssertEqual(t, `<string>Hello, World!</string>`, string(ctx.Fasthttp.Response.Body()))
	ctx.Format(Map{})
	utils.AssertEqual(t, "map[]", string(ctx.Fasthttp.Response.Body()))

	type broken string
	ctx.Fasthttp.Request.Header.Set(HeaderAccept, "broken/accept")
	ctx.Format(broken("Hello, World!"))
	utils.AssertEqual(t, `Hello, World!`, string(ctx.Fasthttp.Response.Body()))
}

// go test -v -run=^$ -bench=Benchmark_Ctx_Format -benchmem -count=4
func Benchmark_Ctx_Format(b *testing.B) {
	app := New()
	c := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(c)
	c.Fasthttp.Request.Header.Set("Accept", "text/plain")
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		c.Format("Hello, World!")
	}
	utils.AssertEqual(b, `Hello, World!`, string(c.Fasthttp.Response.Body()))
}

// go test -v -run=^$ -bench=Benchmark_Ctx_Format_HTML -benchmem -count=4
func Benchmark_Ctx_Format_HTML(b *testing.B) {
	app := New()
	c := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(c)
	c.Fasthttp.Request.Header.Set("Accept", "text/html")
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		c.Format("Hello, World!")
	}
	utils.AssertEqual(b, "<p>Hello, World!</p>", string(c.Fasthttp.Response.Body()))
}

// go test -v -run=^$ -bench=Benchmark_Ctx_Format_JSON -benchmem -count=4
func Benchmark_Ctx_Format_JSON(b *testing.B) {
	app := New()
	c := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(c)
	c.Fasthttp.Request.Header.Set("Accept", "application/json")
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		c.Format("Hello, World!")
	}
	utils.AssertEqual(b, `"Hello, World!"`, string(c.Fasthttp.Response.Body()))
}

// go test -v -run=^$ -bench=Benchmark_Ctx_Format_XML -benchmem -count=4
func Benchmark_Ctx_Format_XML(b *testing.B) {
	app := New()
	c := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(c)
	c.Fasthttp.Request.Header.Set("Accept", "application/xml")
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		c.Format("Hello, World!")
	}
	utils.AssertEqual(b, `<string>Hello, World!</string>`, string(c.Fasthttp.Response.Body()))
}

// go test -run TestContextFormFile
func TestContextFormFile(t *testing.T) {
	// TODO: We should clean this up
	t.Parallel()
	app := New()

	app.POST("/test", func(c *Context) {
		fh, err := c.FormFile("file")
		utils.AssertEqual(t, nil, err)
		utils.AssertEqual(t, "test", fh.Filename)

		f, err := fh.Open()
		utils.AssertEqual(t, nil, err)

		b := new(bytes.Buffer)
		_, err = io.Copy(b, f)
		utils.AssertEqual(t, nil, err)

		f.Close()
		utils.AssertEqual(t, "hello world", b.String())
	})

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	ioWriter, err := writer.CreateFormFile("file", "test")
	utils.AssertEqual(t, nil, err)

	_, err = ioWriter.Write([]byte("hello world"))
	utils.AssertEqual(t, nil, err)

	writer.Close()

	req := httptest.NewRequest("POST", "/test", body)
	req.Header.Set(HeaderContentType, writer.FormDataContentType())
	req.Header.Set(HeaderContentLength, strconv.Itoa(len(body.Bytes())))

	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, StatusOK, resp.StatusCode, "Status code")
}

// go test -run TestContextFormValue
func TestContextFormValue(t *testing.T) {
	t.Parallel()
	app := New()

	app.POST("/test", func(c *Context) {
		utils.AssertEqual(t, "john", c.FormValue("name"))
	})

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	utils.AssertEqual(t, nil, writer.WriteField("name", "john"))

	writer.Close()
	req := httptest.NewRequest("POST", "/test", body)
	req.Header.Set("Content-Type", fmt.Sprintf("multipart/form-data; boundary=%s", writer.Boundary()))
	req.Header.Set("Content-Length", strconv.Itoa(len(body.Bytes())))

	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, StatusOK, resp.StatusCode, "Status code")
}

// go test -run TestContextFresh
func TestContextFresh(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)
	utils.AssertEqual(t, false, ctx.Fresh())

	ctx.Fasthttp.Request.Header.Set(HeaderIfNoneMatch, "*")
	ctx.Fasthttp.Request.Header.Set(HeaderCacheControl, "no-cache")
	utils.AssertEqual(t, false, ctx.Fresh())

	ctx.Fasthttp.Request.Header.Set(HeaderIfNoneMatch, "675af34563dc-tr34")
	ctx.Fasthttp.Request.Header.Set(HeaderCacheControl, "public")
	utils.AssertEqual(t, false, ctx.Fresh())

	ctx.Fasthttp.Request.Header.Set(HeaderIfNoneMatch, "a, b")
	ctx.Fasthttp.Response.Header.Set(HeaderETag, "c")
	utils.AssertEqual(t, false, ctx.Fresh())

	ctx.Fasthttp.Response.Header.Set(HeaderETag, "a")
	utils.AssertEqual(t, true, ctx.Fresh())

	ctx.Fasthttp.Request.Header.Set(HeaderIfModifiedSince, "xxWed, 21 Oct 2015 07:28:00 GMT")
	ctx.Fasthttp.Response.Header.Set(HeaderLastModified, "xxWed, 21 Oct 2015 07:28:00 GMT")
	utils.AssertEqual(t, false, ctx.Fresh())

	ctx.Fasthttp.Response.Header.Set(HeaderLastModified, "Wed, 21 Oct 2015 07:28:00 GMT")
	utils.AssertEqual(t, false, ctx.Fresh())

	ctx.Fasthttp.Request.Header.Set(HeaderIfModifiedSince, "Wed, 21 Oct 2015 07:28:00 GMT")
	utils.AssertEqual(t, false, ctx.Fresh())
}

// go test -run TestContextGet
func TestContextGet(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)
	ctx.Fasthttp.Request.Header.Set(HeaderAcceptCharset, "utf-8, iso-8859-1;q=0.5")
	ctx.Fasthttp.Request.Header.Set(HeaderReferer, "Monster")
	utils.AssertEqual(t, "utf-8, iso-8859-1;q=0.5", ctx.Get(HeaderAcceptCharset))
	utils.AssertEqual(t, "Monster", ctx.Get(HeaderReferer))
	utils.AssertEqual(t, "default", ctx.Get("unknown", "default"))
}

// go test -run TestContextHostname
func TestContextHostname(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)
	ctx.Fasthttp.Request.SetRequestURI("http://google.com/test")
	utils.AssertEqual(t, "google.com", ctx.Hostname())
}

// go test -run TestContextIP
func TestContextIP(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)
	utils.AssertEqual(t, "0.0.0.0", ctx.IP())
}

// go test -run TestContextIPs
func TestContextIPs(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)
	ctx.Fasthttp.Request.Header.Set(HeaderXForwardedFor, "127.0.0.1, 127.0.0.1, 127.0.0.1")
	utils.AssertEqual(t, []string{"127.0.0.1", "127.0.0.1", "127.0.0.1"}, ctx.IPs())
}

// go test -v -run=^$ -bench=Benchmark_Ctx_IPs -benchmem -count=4
func Benchmark_Ctx_IPs(b *testing.B) {
	app := New()
	c := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(c)
	c.Fasthttp.Request.Header.Set(HeaderXForwardedFor, "127.0.0.1, 127.0.0.1, 127.0.0.1")
	var res []string
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		res = c.IPs()
	}
	utils.AssertEqual(b, []string{"127.0.0.1", "127.0.0.1", "127.0.0.1"}, res)
}

// go test -run TestContextIs
func TestContextIs(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)
	ctx.Fasthttp.Request.Header.Set(HeaderContentType, MIMETextHTML+"; boundary=something")
	utils.AssertEqual(t, true, ctx.Is(".html"))
	utils.AssertEqual(t, true, ctx.Is("html"))
	utils.AssertEqual(t, false, ctx.Is("json"))
	utils.AssertEqual(t, false, ctx.Is(".json"))
	utils.AssertEqual(t, false, ctx.Is(""))
	utils.AssertEqual(t, false, ctx.Is(".foooo"))

	ctx.Fasthttp.Request.Header.Set(HeaderContentType, MIMEApplicationJSONCharsetUTF8)
	utils.AssertEqual(t, false, ctx.Is("html"))
	utils.AssertEqual(t, true, ctx.Is("json"))
	utils.AssertEqual(t, true, ctx.Is(".json"))

	ctx.Fasthttp.Request.Header.Set(HeaderContentType, " application/json;charset=UTF-8")
	utils.AssertEqual(t, false, ctx.Is("html"))
	utils.AssertEqual(t, true, ctx.Is("json"))
	utils.AssertEqual(t, true, ctx.Is(".json"))

	ctx.Fasthttp.Request.Header.Set(HeaderContentType, MIMEApplicationXMLCharsetUTF8)
	utils.AssertEqual(t, false, ctx.Is("html"))
	utils.AssertEqual(t, true, ctx.Is("xml"))
	utils.AssertEqual(t, true, ctx.Is(".xml"))

	ctx.Fasthttp.Request.Header.Set(HeaderContentType, MIMETextPlain)
	utils.AssertEqual(t, false, ctx.Is("html"))
	utils.AssertEqual(t, true, ctx.Is("txt"))
	utils.AssertEqual(t, true, ctx.Is(".txt"))
}

// go test -v -run=^$ -bench=Benchmark_Ctx_Is -benchmem -count=4
func Benchmark_Ctx_Is(b *testing.B) {
	app := New()
	c := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(c)
	c.Fasthttp.Request.Header.Set(HeaderContentType, MIMEApplicationJSON)
	var res bool
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = c.Is(".json")
		res = c.Is("json")
	}
	utils.AssertEqual(b, true, res)
}

// go test -run TestContextLocals
func TestContextLocals(t *testing.T) {
	app := New()
	app.Use(func(c *Context) {
		c.Locals("john", "doe")
		c.Next()
	})
	app.GET("/test", func(c *Context) {
		utils.AssertEqual(t, "doe", c.Locals("john"))
	})
	resp, err := app.Test(httptest.NewRequest("GET", "/test", nil))
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, StatusOK, resp.StatusCode, "Status code")
}

// go test -run TestContextMethod
func TestContextMethod(t *testing.T) {
	t.Parallel()
	fctx := &fasthttp.RequestCtx{}
	fctx.Request.Header.SetMethod("GET")
	app := New()
	ctx := app.AcquireContext(fctx)
	defer app.ReleaseContext(ctx)
	utils.AssertEqual(t, "GET", ctx.Method())
	ctx.Method("POST")
	utils.AssertEqual(t, "POST", ctx.Method())

	ctx.Method("MethodInvalid")
	utils.AssertEqual(t, "POST", ctx.Method())
}

// go test -run TestContextMultipartForm
func TestContextMultipartForm(t *testing.T) {
	t.Parallel()
	app := New()

	app.POST("/test", func(c *Context) {
		result, err := c.MultipartForm()
		utils.AssertEqual(t, nil, err)
		utils.AssertEqual(t, "john", result.Value["name"][0])
	})

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	utils.AssertEqual(t, nil, writer.WriteField("name", "john"))

	writer.Close()
	req := httptest.NewRequest("POST", "/test", body)
	req.Header.Set(HeaderContentType, fmt.Sprintf("multipart/form-data; boundary=%s", writer.Boundary()))
	req.Header.Set(HeaderContentLength, strconv.Itoa(len(body.Bytes())))

	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, StatusOK, resp.StatusCode, "Status code")
}

// go test -run TestContextOriginalURL
func TestContextOriginalURL(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)
	ctx.Fasthttp.Request.Header.SetRequestURI("http://google.com/test?search=demo")
	utils.AssertEqual(t, "http://google.com/test?search=demo", ctx.OriginalURL())
}

// go test -race -run TestContextParams
func TestContextParams(t *testing.T) {
	t.Parallel()
	app := New()
	app.GET("/test/:user", func(c *Context) {
		utils.AssertEqual(t, "john", c.Params("user"))
	})
	app.GET("/test2/*", func(c *Context) {
		utils.AssertEqual(t, "im/a/cookie", c.Params("*"))
	})
	app.GET("/test3/:optional?", func(c *Context) {
		utils.AssertEqual(t, "", c.Params("optional"))
	})
	resp, err := app.Test(httptest.NewRequest("GET", "/test/john", nil))
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, StatusOK, resp.StatusCode, "Status code")
	resp, err = app.Test(httptest.NewRequest("GET", "/test2/im/a/cookie", nil))
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, StatusOK, resp.StatusCode, "Status code")
	resp, err = app.Test(httptest.NewRequest("GET", "/test3", nil))
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, StatusOK, resp.StatusCode, "Status code")
}

// go test -v -run=^$ -bench=Benchmark_Ctx_Params -benchmem -count=4
func Benchmark_Ctx_Params(b *testing.B) {
	app := New()
	c := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(c)
	c.route = &Route{
		Params: []string{
			"param1", "param2", "param3", "param4",
		},
	}
	c.values = []string{
		"john", "doe", "is", "awesome",
	}
	var res string
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = c.Params("param1")
		_ = c.Params("param2")
		_ = c.Params("param3")
		res = c.Params("param4")
	}
	utils.AssertEqual(b, "awesome", res)
}

// go test -run TestContextPath
func TestContextPath(t *testing.T) {
	t.Parallel()
	app := New()
	app.GET("/test/:user", func(c *Context) {
		utils.AssertEqual(t, "/test/john", c.Path())
		// not strict && case insensitive
		utils.AssertEqual(t, "/ABC/", c.Path("/ABC/"))
		utils.AssertEqual(t, "/test/john/", c.Path("/test/john/"))
	})
	resp, err := app.Test(httptest.NewRequest("GET", "/test/john", nil))
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, StatusOK, resp.StatusCode, "Status code")
}

// go test -run TestContextProtocol
func TestContextProtocol(t *testing.T) {
	app := New()

	freq := &fasthttp.RequestCtx{}
	freq.Request.Header.Set("X-Forwarded", "invalid")

	ctx := app.AcquireContext(freq)
	defer app.ReleaseContext(ctx)
	ctx.Fasthttp.Request.Header.Set(HeaderXForwardedProto, "https")
	utils.AssertEqual(t, "https", ctx.Protocol())
	ctx.Fasthttp.Request.Header.Reset()

	ctx.Fasthttp.Request.Header.Set(HeaderXForwardedProtocol, "https")
	utils.AssertEqual(t, "https", ctx.Protocol())
	ctx.Fasthttp.Request.Header.Reset()

	ctx.Fasthttp.Request.Header.Set(HeaderXForwardedSsl, "on")
	utils.AssertEqual(t, "https", ctx.Protocol())
	ctx.Fasthttp.Request.Header.Reset()

	ctx.Fasthttp.Request.Header.Set(HeaderXUrlScheme, "https")
	utils.AssertEqual(t, "https", ctx.Protocol())
	ctx.Fasthttp.Request.Header.Reset()

	utils.AssertEqual(t, "http", ctx.Protocol())
}

// go test -v -run=^$ -bench=Benchmark_Ctx_Protocol -benchmem -count=4
func Benchmark_Ctx_Protocol(b *testing.B) {
	app := New()
	c := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(c)
	var res string
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		res = c.Protocol()
	}
	utils.AssertEqual(b, "http", res)
}

// go test -run TestContextQuery
func TestContextQuery(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)
	ctx.Fasthttp.Request.URI().SetQueryString("search=john&age=20")
	utils.AssertEqual(t, "john", ctx.Query("search"))
	utils.AssertEqual(t, "20", ctx.Query("age"))
	utils.AssertEqual(t, "default", ctx.Query("unknown", "default"))
}

// go test -run TestContextRange
func TestContextRange(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)

	var (
		result Range
		err    error
	)

	result, err = ctx.Range(1000)
	utils.AssertEqual(t, true, err != nil)

	ctx.Fasthttp.Request.Header.Set(HeaderRange, "bytes=500")
	result, err = ctx.Range(1000)
	utils.AssertEqual(t, true, err != nil)

	ctx.Fasthttp.Request.Header.Set(HeaderRange, "bytes=500=")
	result, err = ctx.Range(1000)
	utils.AssertEqual(t, true, err != nil)

	ctx.Fasthttp.Request.Header.Set(HeaderRange, "bytes=500-300")
	result, err = ctx.Range(1000)
	utils.AssertEqual(t, true, err != nil)

	testRange := func(header string, start, end int) {
		ctx.Fasthttp.Request.Header.Set(HeaderRange, header)
		result, err = ctx.Range(1000)
		utils.AssertEqual(t, nil, err)
		utils.AssertEqual(t, "bytes", result.Type)
		utils.AssertEqual(t, start, result.Ranges[0].Start)
		utils.AssertEqual(t, end, result.Ranges[0].End)
	}

	testRange("bytes=a-700", 300, 999)
	testRange("bytes=500-b", 500, 999)
	testRange("bytes=500-1000", 500, 999)
	testRange("bytes=500-700", 500, 700)
}

// go test -run TestContextRoute
func TestContextRoute(t *testing.T) {
	t.Parallel()
	app := New()
	app.GET("/test", func(c *Context) {
		utils.AssertEqual(t, "/test", c.Route().Path)
	})
	resp, err := app.Test(httptest.NewRequest("GET", "/test", nil))
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, StatusOK, resp.StatusCode, "Status code")

	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)

	utils.AssertEqual(t, "/", ctx.Route().Path)
	utils.AssertEqual(t, "GET", ctx.Route().Method)
	utils.AssertEqual(t, 0, len(ctx.Route().Handlers))
}

// go test -run TestContextRouteNormalized
func TestContextRouteNormalized(t *testing.T) {
	t.Parallel()
	app := New()
	app.GET("/test", func(c *Context) {
		utils.AssertEqual(t, "/test", c.Route().Path)
	})
	resp, err := app.Test(httptest.NewRequest("GET", "//test", nil))
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, StatusNotFound, resp.StatusCode, "Status code")
}

// go test -run TestContextSaveFile
func TestContextSaveFile(t *testing.T) {
	// TODO We should clean this up
	t.Parallel()
	app := New()

	app.POST("/test", func(c *Context) {
		fh, err := c.FormFile("file")
		utils.AssertEqual(t, nil, err)

		tempFile, err := ioutil.TempFile(os.TempDir(), "test-")
		utils.AssertEqual(t, nil, err)

		defer os.Remove(tempFile.Name())
		err = c.SaveFile(fh, tempFile.Name())
		utils.AssertEqual(t, nil, err)

		bs, err := ioutil.ReadFile(tempFile.Name())
		utils.AssertEqual(t, nil, err)
		utils.AssertEqual(t, "hello world", string(bs))
	})

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	ioWriter, err := writer.CreateFormFile("file", "test")
	utils.AssertEqual(t, nil, err)

	_, err = ioWriter.Write([]byte("hello world"))
	utils.AssertEqual(t, nil, err)
	writer.Close()

	req := httptest.NewRequest("POST", "/test", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Content-Length", strconv.Itoa(len(body.Bytes())))

	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, StatusOK, resp.StatusCode, "Status code")
}

// go test -run TestContextSecure
func TestContextSecure(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)
	// TODO Add TLS conn
	utils.AssertEqual(t, false, ctx.Secure())
}

// go test -run TestContextStale
func TestContextStale(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)
	utils.AssertEqual(t, true, ctx.Stale())
}

// go test -run TestContextSubdomains
func TestContextSubdomains(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)
	ctx.Fasthttp.Request.URI().SetHost("john.doe.is.awesome.google.com")
	utils.AssertEqual(t, []string{"john", "doe"}, ctx.Subdomains(4))

	ctx.Fasthttp.Request.URI().SetHost("localhost:3000")
	utils.AssertEqual(t, []string{"localhost:3000"}, ctx.Subdomains())
}

// go test -v -run=^$ -bench=Benchmark_Ctx_Subdomains -benchmem -count=4
func Benchmark_Ctx_Subdomains(b *testing.B) {
	app := New()
	c := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(c)
	c.Fasthttp.Request.SetRequestURI("http://john.doe.google.com")
	var res []string
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		res = c.Subdomains()
	}
	utils.AssertEqual(b, []string{"john", "doe"}, res)
}

// go test -run TestContextClearCookie
func TestContextClearCookie(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)
	ctx.Fasthttp.Request.Header.Set(HeaderCookie, "john=doe")
	ctx.ClearCookie("john")
	utils.AssertEqual(t, true, strings.HasPrefix(string(ctx.Fasthttp.Response.Header.Peek(HeaderSetCookie)), "john=; expires="))

	ctx.Fasthttp.Request.Header.Set(HeaderCookie, "test1=dummy")
	ctx.Fasthttp.Request.Header.Set(HeaderCookie, "test2=dummy")
	ctx.ClearCookie()
	utils.AssertEqual(t, true, strings.Contains(string(ctx.Fasthttp.Response.Header.Peek(HeaderSetCookie)), "test1=; expires="))
	utils.AssertEqual(t, true, strings.Contains(string(ctx.Fasthttp.Response.Header.Peek(HeaderSetCookie)), "test2=; expires="))
}

// go test -race -run TestContextDownload
func TestContextDownload(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)

	ctx.Download("context.go", "Awesome File!")

	f, err := os.Open("./context.go")
	utils.AssertEqual(t, nil, err)
	defer f.Close()

	expect, err := ioutil.ReadAll(f)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, expect, ctx.Fasthttp.Response.Body())
	utils.AssertEqual(t, `attachment; filename="Awesome+File%21"`, string(ctx.Fasthttp.Response.Header.Peek(HeaderContentDisposition)))
}

// go test -race -run TestContextSendFile
func TestContextSendFile(t *testing.T) {
	t.Parallel()
	app := New()

	// fetch file content
	f, err := os.Open("./context.go")
	utils.AssertEqual(t, nil, err)
	defer f.Close()
	expectFileContent, err := ioutil.ReadAll(f)
	utils.AssertEqual(t, nil, err)
	// fetch file info for the not modified test case
	fI, err := os.Stat("./context.go")
	utils.AssertEqual(t, nil, err)

	// simple test case
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	err = ctx.SendFile("context.go")
	// check expectation
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, expectFileContent, ctx.Fasthttp.Response.Body())
	utils.AssertEqual(t, StatusOK, ctx.Fasthttp.Response.StatusCode())
	app.ReleaseContext(ctx)

	// test with custom error code
	ctx = app.AcquireContext(&fasthttp.RequestCtx{})
	err = ctx.Status(StatusInternalServerError).SendFile("context.go")
	// check expectation
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, expectFileContent, ctx.Fasthttp.Response.Body())
	utils.AssertEqual(t, StatusInternalServerError, ctx.Fasthttp.Response.StatusCode())
	app.ReleaseContext(ctx)

	// test not modified
	ctx = app.AcquireContext(&fasthttp.RequestCtx{})
	ctx.Fasthttp.Request.Header.Set(HeaderIfModifiedSince, fI.ModTime().Format(time.RFC1123))
	err = ctx.SendFile("context.go")
	// check expectation
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, StatusNotModified, ctx.Fasthttp.Response.StatusCode())
	utils.AssertEqual(t, []byte(nil), ctx.Fasthttp.Response.Body())
	app.ReleaseContext(ctx)

	// test 404
	// ctx = app.AcquireCtx(&fasthttp.RequestCtx{})
	// err = ctx.SendFile("./john_doe.go")
	// // check expectation
	// utils.AssertEqual(t, StatusNotFound, ctx.Fasthttp.Response.StatusCode())
	// app.ReleaseCtx(ctx)
}

// go test -race -run TestContextSendFile_Immutable
func TestContextSendFile_Immutable(t *testing.T) {
	t.Parallel()
	app := New()
	app.GET("/:file", func(c *Context) {
		file := c.Params("file")
		if err := c.SendFile("../test/" + file + ".html"); err != nil {
			utils.AssertEqual(t, nil, err)
		}
		utils.AssertEqual(t, "index", fmt.Sprintf("%s", file))
		c.Send(file)
	})
	// 1st try
	resp, err := app.Test(httptest.NewRequest("GET", "/index", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, StatusOK, resp.StatusCode)
	// 2nd try
	resp, err = app.Test(httptest.NewRequest("GET", "/index", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, StatusOK, resp.StatusCode)
}

// go test -run TestContextJSON
func TestContextJSON(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)

	utils.AssertEqual(t, true, ctx.JSON(complex(1, 1)) != nil)

	ctx.JSON(Map{ // map has no order
		"Name": "Grame",
		"Age":  20,
	})
	utils.AssertEqual(t, `{"Age":20,"Name":"Grame"}`, string(ctx.Fasthttp.Response.Body()))
	utils.AssertEqual(t, "application/json", string(ctx.Fasthttp.Response.Header.Peek("content-type")))

	testEmpty := func(v interface{}, r string) {
		err := ctx.JSON(v)
		utils.AssertEqual(t, nil, err)
		utils.AssertEqual(t, r, string(ctx.Fasthttp.Response.Body()))
	}

	testEmpty(nil, "null")
	testEmpty("", `""`)
	testEmpty(0, "0")
	testEmpty([]int{}, "[]")
}

// go test -v  -run=^$ -bench=Benchmark_Ctx_JSON -benchmem -count=4
func Benchmark_Ctx_JSON(b *testing.B) {
	app := New()
	c := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(c)
	type SomeStruct struct {
		Name string
		Age  uint8
	}
	data := SomeStruct{
		Name: "Grame",
		Age:  20,
	}
	var err error
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		err = c.JSON(data)
	}
	utils.AssertEqual(b, nil, err)
	utils.AssertEqual(b, `{"Name":"Grame","Age":20}`, string(c.Fasthttp.Response.Body()))
}

// go test -run TestContextJSONP
func TestContextJSONP(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)

	utils.AssertEqual(t, true, ctx.JSONP(complex(1, 1)) != nil)

	ctx.JSONP(Map{
		"Name": "Grame",
		"Age":  20,
	})
	utils.AssertEqual(t, `callback({"Age":20,"Name":"Grame"});`, string(ctx.Fasthttp.Response.Body()))
	utils.AssertEqual(t, "application/javascript; charset=utf-8", string(ctx.Fasthttp.Response.Header.Peek("content-type")))

	ctx.JSONP(Map{
		"Name": "Grame",
		"Age":  20,
	}, "john")
	utils.AssertEqual(t, `john({"Age":20,"Name":"Grame"});`, string(ctx.Fasthttp.Response.Body()))
	utils.AssertEqual(t, "application/javascript; charset=utf-8", string(ctx.Fasthttp.Response.Header.Peek("content-type")))
}

// go test -v  -run=^$ -bench=Benchmark_Ctx_JSONP -benchmem -count=4
func Benchmark_Ctx_JSONP(b *testing.B) {
	app := New()
	c := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(c)
	type SomeStruct struct {
		Name string
		Age  uint8
	}
	data := SomeStruct{
		Name: "Grame",
		Age:  20,
	}
	var callback = "emit"
	var err error
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		err = c.JSONP(data, callback)
	}
	utils.AssertEqual(b, nil, err)
	utils.AssertEqual(b, `emit({"Name":"Grame","Age":20});`, string(c.Fasthttp.Response.Body()))
}

// go test -run TestContextLinks
func TestContextLinks(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)

	ctx.Links()
	utils.AssertEqual(t, "", string(ctx.Fasthttp.Response.Header.Peek(HeaderLink)))

	ctx.Links(
		"http://api.example.com/users?page=2", "next",
		"http://api.example.com/users?page=5", "last",
	)
	utils.AssertEqual(t, `<http://api.example.com/users?page=2>; rel="next",<http://api.example.com/users?page=5>; rel="last"`, string(ctx.Fasthttp.Response.Header.Peek(HeaderLink)))
}

// go test -v  -run=^$ -bench=Benchmark_Ctx_Links -benchmem -count=4
func Benchmark_Ctx_Links(b *testing.B) {
	app := New()
	c := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(c)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		c.Links(
			"http://api.example.com/users?page=2", "next",
			"http://api.example.com/users?page=5", "last",
		)
	}
}

// go test -run TestContextLocation
func TestContextLocation(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)
	ctx.Location("http://example.com")
	utils.AssertEqual(t, "http://example.com", string(ctx.Fasthttp.Response.Header.Peek(HeaderLocation)))
}

// go test -run TestContextNext
func TestContextNext(t *testing.T) {
	app := New()
	app.Use("/", func(c *Context) {
		c.Next()
	})
	app.GET("/test", func(c *Context) {
		c.Set("X-Next-Result", "Works")
	})
	resp, err := app.Test(httptest.NewRequest("GET", "http://example.com/test", nil))
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, StatusOK, resp.StatusCode, "Status code")
	utils.AssertEqual(t, "Works", resp.Header.Get("X-Next-Result"))
}

// go test -run TestContextNext_Error
func TestContextNext_Error(t *testing.T) {
	app := New()
	app.Use("/", func(c *Context) {
		c.Set("X-Next-Result", "Works")
		c.Next(ErrNotFound)
	})

	resp, err := app.Test(httptest.NewRequest("GET", "http://example.com/test", nil))
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, StatusNotFound, resp.StatusCode, "Status code")
	utils.AssertEqual(t, "Works", resp.Header.Get("X-Next-Result"))
}

// go test -run TestContextRedirect
func TestContextRedirect(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)

	ctx.Redirect("http://default.com")
	utils.AssertEqual(t, 302, ctx.Fasthttp.Response.StatusCode())
	utils.AssertEqual(t, "http://default.com", string(ctx.Fasthttp.Response.Header.Peek(HeaderLocation)))

	ctx.Redirect("http://example.com", 301)
	utils.AssertEqual(t, 301, ctx.Fasthttp.Response.StatusCode())
	utils.AssertEqual(t, "http://example.com", string(ctx.Fasthttp.Response.Header.Peek(HeaderLocation)))
}

// go test -run TestContextRender
func TestContextRender(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)
	err := ctx.Render("../test/TEST_DATA/template.html", Map{
		"Title": "Hello, World!",
	})

	buf := bytebufferpool.Get()
	_, _ = buf.WriteString("overwrite")
	defer bytebufferpool.Put(buf)

	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, "<h1>Hello, World!</h1>", string(ctx.Fasthttp.Response.Body()))

	err = ctx.Render("../test/TEST_DATA/invalid.html", nil)
	utils.AssertEqual(t, false, err == nil)
}

type testTemplateEngine struct {
	mu        sync.Mutex
	templates *template.Template
}

func (t *testTemplateEngine) Render(w io.Writer, name string, bind interface{}, layout ...string) error {
	return t.templates.ExecuteTemplate(w, name, bind)
}

func (t *testTemplateEngine) Load() error {
	t.templates = template.Must(template.ParseGlob("../test/TEST_DATA/*.tmpl"))
	return nil
}

// go test -run TestContextRender_Engine
func TestContextRender_Engine(t *testing.T) {
	engine := &testTemplateEngine{}
	engine.Load()
	app := New()
	app.Settings.Views = engine
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)
	err := ctx.Render("index.tmpl", Map{
		"Title": "Hello, World!",
	})
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, "<h1>Hello, World!</h1>", string(ctx.Fasthttp.Response.Body()))
}

func Benchmark_Ctx_Render_Engine(b *testing.B) {
	engine := &testTemplateEngine{}
	err := engine.Load()
	utils.AssertEqual(b, nil, err)
	app := New()
	app.Settings.Views = engine
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		err = ctx.Render("index.tmpl", Map{
			"Title": "Hello, World!",
		})
	}
	utils.AssertEqual(b, nil, err)
	utils.AssertEqual(b, "<h1>Hello, World!</h1>", string(ctx.Fasthttp.Response.Body()))
}

// go test -run TestContextRender_Go_Template
func TestContextRender_Go_Template(t *testing.T) {
	t.Parallel()
	file, err := ioutil.TempFile(os.TempDir(), "fiber")
	utils.AssertEqual(t, nil, err)
	defer os.Remove(file.Name())
	_, err = file.Write([]byte("template"))
	utils.AssertEqual(t, nil, err)
	err = file.Close()
	utils.AssertEqual(t, nil, err)
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)
	err = ctx.Render(file.Name(), nil)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, "template", string(ctx.Fasthttp.Response.Body()))
}

// go test -run TestContextSend
func TestContextSend(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)
	ctx.Send([]byte("Hello, World"))
	ctx.Send("Don't crash please")
	ctx.Send(1337)
	utils.AssertEqual(t, "1337", string(ctx.Fasthttp.Response.Body()))
}

// go test -v  -run=^$ -bench=Benchmark_Ctx_Send -benchmem -count=4
func Benchmark_Ctx_Send(b *testing.B) {
	app := New()
	c := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(c)
	var str = "Hello, World!"
	var byt = []byte("Hello, World!")
	var nmb = 123
	var bol = true
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		c.Send(str)
		c.Send(byt)
		c.Send(nmb)
		c.Send(bol)
	}
	utils.AssertEqual(b, "true", string(c.Fasthttp.Response.Body()))
}

// go test -run TestContextSendBytes
func TestContextSendBytes(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)
	ctx.SendBytes([]byte("Hello, World!"))
	utils.AssertEqual(t, "Hello, World!", string(ctx.Fasthttp.Response.Body()))
}

// go test -v  -run=^$ -bench=Benchmark_Ctx_Benchmark_Ctx_SendBytes -benchmem -count=4
func Benchmark_Ctx_SendBytes(b *testing.B) {
	app := New()
	c := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(c)
	var byt = []byte("Hello, World!")
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		c.SendBytes(byt)
	}
	utils.AssertEqual(b, "Hello, World!", string(c.Fasthttp.Response.Body()))
}

// go test -run TestContextSendStatus
func TestContextSendStatus(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)
	ctx.SendStatus(415)
	utils.AssertEqual(t, 415, ctx.Fasthttp.Response.StatusCode())
	utils.AssertEqual(t, "Unsupported Media Type", string(ctx.Fasthttp.Response.Body()))
}

// go test -run TestContextSendString
func TestContextSendString(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)
	ctx.SendString("Don't crash please")
	utils.AssertEqual(t, "Don't crash please", string(ctx.Fasthttp.Response.Body()))
}

// go test -run TestContextSendStream
func TestContextSendStream(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)

	ctx.SendStream(bytes.NewReader([]byte("Don't crash please")))
	utils.AssertEqual(t, "Don't crash please", string(ctx.Fasthttp.Response.Body()))

	ctx.SendStream(bytes.NewReader([]byte("Don't crash please")), len([]byte("Don't crash please")))
	utils.AssertEqual(t, "Don't crash please", string(ctx.Fasthttp.Response.Body()))

	ctx.SendStream(bufio.NewReader(bytes.NewReader([]byte("Hello bufio"))))
	utils.AssertEqual(t, "Hello bufio", string(ctx.Fasthttp.Response.Body()))

	file, err := os.Open("../test/index.html")
	utils.AssertEqual(t, nil, err)
	ctx.SendStream(bufio.NewReader(file))
	utils.AssertEqual(t, true, (ctx.Fasthttp.Response.Header.ContentLength() > 200))
}

// go test -run TestContextSet
func TestContextSet(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)
	ctx.Set("X-1", "1")
	ctx.Set("X-2", "2")
	ctx.Set("X-3", "3")
	ctx.Set("X-3", "1337")
	utils.AssertEqual(t, "1", string(ctx.Fasthttp.Response.Header.Peek("x-1")))
	utils.AssertEqual(t, "2", string(ctx.Fasthttp.Response.Header.Peek("x-2")))
	utils.AssertEqual(t, "1337", string(ctx.Fasthttp.Response.Header.Peek("x-3")))
}

// go test -run TestContextStatus
func TestContextStatus(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)
	ctx.Status(400)
	utils.AssertEqual(t, 400, ctx.Fasthttp.Response.StatusCode())
	ctx.Status(415).Send("Hello, World")
	utils.AssertEqual(t, 415, ctx.Fasthttp.Response.StatusCode())
	utils.AssertEqual(t, "Hello, World", string(ctx.Fasthttp.Response.Body()))
}

// go test -run TestContextType
func TestContextType(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)
	ctx.Type(".json")
	utils.AssertEqual(t, "application/json", string(ctx.Fasthttp.Response.Header.Peek("Content-Type")))

	ctx.Type("json", "utf-8")
	utils.AssertEqual(t, "application/json; charset=utf-8", string(ctx.Fasthttp.Response.Header.Peek("Content-Type")))

	ctx.Type(".html")
	utils.AssertEqual(t, "text/html", string(ctx.Fasthttp.Response.Header.Peek("Content-Type")))

	ctx.Type("html", "utf-8")
	utils.AssertEqual(t, "text/html; charset=utf-8", string(ctx.Fasthttp.Response.Header.Peek("Content-Type")))
}

// go test -v  -run=^$ -bench=Benchmark_Ctx_Type -benchmem -count=4
func Benchmark_Ctx_Type(b *testing.B) {
	app := New()
	c := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(c)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		c.Type(".json")
		c.Type("json")
	}
}

// go test -v  -run=^$ -bench=Benchmark_Ctx_Type_Charset -benchmem -count=4
func Benchmark_Ctx_Type_Charset(b *testing.B) {
	app := New()
	c := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(c)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		c.Type(".json", "utf-8")
		c.Type("json", "utf-8")
	}
}

// go test -run TestContextVary
func TestContextVary(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)
	ctx.Vary("Origin")
	ctx.Vary("User-Agent")
	ctx.Vary("Accept-Encoding", "Accept")
	utils.AssertEqual(t, "Origin, User-Agent, Accept-Encoding, Accept", string(ctx.Fasthttp.Response.Header.Peek("Vary")))
}

// go test -v  -run=^$ -bench=Benchmark_Ctx_Vary -benchmem -count=4
func Benchmark_Ctx_Vary(b *testing.B) {
	app := New()
	c := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(c)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		c.Vary("Origin", "User-Agent")
	}
}

// go test -run TestContextWrite
func TestContextWrite(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)
	ctx.Write("Hello, ")
	ctx.Write([]byte("World! "))
	ctx.Write(123)
	ctx.Write(123.321)
	ctx.Write(true)
	ctx.Write(bytes.NewReader([]byte("Don't crash please")))
	utils.AssertEqual(t, "Don't crash please", string(ctx.Fasthttp.Response.Body()))
}

// go test -v -run=^$ -bench=Benchmark_Ctx_Write -benchmem -count=4
func Benchmark_Ctx_Write(b *testing.B) {
	app := New()
	c := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(c)
	var str = "Hello, World!"
	var byt = []byte("Hello, World!")
	var nmb = 123
	var bol = true
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		c.Write(str)
		c.Write(byt)
		c.Write(nmb)
		c.Write(bol)
	}
}

// go test -run TestContextXHR
func TestContextXHR(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)
	ctx.Fasthttp.Request.Header.Set(HeaderXRequestedWith, "XMLHttpRequest")
	utils.AssertEqual(t, true, ctx.XHR())
}

// go test -v  -run=^$ -bench=Benchmark_Ctx_SendString_B -benchmem -count=4
func Benchmark_Ctx_SendString_B(b *testing.B) {
	app := New()
	c := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(c)
	body := "Hello, world!"
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		c.SendString(body)
	}
	utils.AssertEqual(b, []byte("Hello, world!"), c.Fasthttp.Response.Body())
}

// go test -v  -run=^$ -bench=Benchmark_Ctx_SendBytes_B -benchmem -count=4
func Benchmark_Ctx_SendBytes_B(b *testing.B) {
	app := New()
	c := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(c)
	body := []byte("Hello, world!")
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		c.SendBytes(body)
	}
	utils.AssertEqual(b, []byte("Hello, world!"), c.Fasthttp.Response.Body())
}

// go test -run TestContextQueryParser
func TestContextQueryParser(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)
	type Query struct {
		ID    int
		Name  string
		Hobby []string
	}
	ctx.Fasthttp.Request.SetBody([]byte(``))
	ctx.Fasthttp.Request.Header.SetContentType("")
	ctx.Fasthttp.Request.URI().SetQueryString("id=1&name=tom&hobby=basketball&hobby=football")
	q := new(Query)
	utils.AssertEqual(t, nil, ctx.QueryParser(q))
	utils.AssertEqual(t, 2, len(q.Hobby))

	empty := new(Query)
	ctx.Fasthttp.Request.URI().SetQueryString("")
	utils.AssertEqual(t, nil, ctx.QueryParser(empty))
	utils.AssertEqual(t, 0, len(empty.Hobby))
}

// go test -v  -run=^$ -bench=Benchmark_Ctx_QueryParser -benchmem -count=4
func Benchmark_Ctx_QueryParser(b *testing.B) {
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)
	type Query struct {
		ID    int
		Name  string
		Hobby []string
	}
	ctx.Fasthttp.Request.SetBody([]byte(``))
	ctx.Fasthttp.Request.Header.SetContentType("")
	ctx.Fasthttp.Request.URI().SetQueryString("id=1&name=tom&hobby=basketball&hobby=football")
	q := new(Query)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		ctx.QueryParser(q)
	}
	utils.AssertEqual(b, nil, ctx.QueryParser(q))
}

// go test -run TestContextError
func TestContextError(t *testing.T) {
	t.Parallel()
	app := New()
	ctx := app.AcquireContext(&fasthttp.RequestCtx{})
	defer app.ReleaseContext(ctx)
	ctx.Next(errors.New("Hi I'm an error"))
	utils.AssertEqual(t, "Hi I'm an error", ctx.Error().Error())
}
