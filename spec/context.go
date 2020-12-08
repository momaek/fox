package engine

import (
	"mime/multipart"
	"net/http"
)

// ResponseWriter ...
type ResponseWriter interface {
	http.ResponseWriter
	http.Hijacker
	http.Flusher
	http.CloseNotifier

	// Returns the HTTP response status code of the current request.
	Status() int

	// Returns the number of bytes already written into the response http body.
	// See Written()
	Size() int

	// Writes the string into the response body.
	WriteString(string) (int, error)

	// Returns true if the response body was already written.
	Written() bool

	// Forces to write the http header (status code + headers).
	WriteHeaderNow()

	// get the http.Pusher for server push
	Pusher() http.Pusher
}

// RequestIDKey header request id key
var RequestIDKey = "x-request-id"

// Context context interface
type Context interface {

	// Copy returns a copy of the current context that can be safely used outside the request's scope.
	Copy() Context

	// Request with http.Request
	Request() *http.Request

	// Writer with ResponseWriter
	Writer() ResponseWriter

	// RequestID return request id
	RequestID() string

	// GetLogger return logger
	GetLogger() Logger

	// FullPath returns a matched route full path. For not found routes
	// returns an empty string.
	//     router.GET("/user/:id", func(c *gin.Context) {
	//         c.FullPath() == "/user/:id" // true
	//     })
	FullPath() string

	// * FLOW CONTROL
	// ******************************************************************

	// Abort actively Abort all subsequent handler executions, but the current handler needs to actively return
	Abort()

	// * METADATA MANAGEMENT
	// ******************************************************************

	// Set is used to store a new key/value pair exclusively for this context.
	Set(key string, value interface{})

	// Get returns the value for the given key, ie: (value, true).
	Get(key string) (value interface{}, exists bool)

	// MustGet returns the value for the given key if it exists, otherwise it panics.
	MustGet(key string) interface{}

	// * HEADER
	// ******************************************************************

	// ContentType returns the Content-Type header of the request.
	ContentType() string

	// ClientIP return client IP
	ClientIP() string

	// * PATH
	// ******************************************************************

	// Param return URL path parameter value
	// GET /user/:id
	// GET /user/qiniu	c.Param("id") => "qiniu"
	// GET /user/12345	c.Param("id") => "12345"
	Param(key string) (value string)

	// * QUERY
	// ******************************************************************

	// Query return url query parameter value
	// 	GET /path?id=1234&name=Manu&value=
	// 		c.Query("id") == "1234"
	// 		c.Query("name") == "Manu"
	// 		c.Query("value") == ""
	// 		c.Query("wtf") == ""
	Query(key string) (value string)

	// GetQuery is like Query(), it returns the keyed url query value
	// if it exists `(value, true)` (even when the value is an empty string),
	// otherwise it returns `("", false)`.
	// It is shortcut for `c.Request.URL.Query().Get(key)`
	//     GET /?name=Manu&lastname=
	//     ("Manu", true) == c.GetQuery("name")
	//     ("", false) == c.GetQuery("id")
	//     ("", true) == c.GetQuery("lastname")
	GetQuery(key string) (value string, exists bool)

	// QueryArray returns a slice of strings for a given query key.
	// The length of the slice depends on the number of params with the given key.
	QueryArray(key string) []string

	// QueryMap returns a map for a given query key.
	QueryMap(key string) map[string]string

	// GetQueryMap returns a map for a given query key, plus a boolean value
	// whether at least one value exists for the given key.
	GetQueryMap(key string) (map[string]string, bool)

	// ShouldBindURI from url path
	ShouldBindURI(obj interface{}) error

	// BindQuery bind GET request or application/x-www-form-urlencoded, multipart/form-data
	BindQuery(obj interface{}) error

	// ShouldBindQuery must bind query
	ShouldBindQuery(obj interface{}) error

	// * FORM
	// ******************************************************************

	// PostForm returns the specified key from a POST urlencoded form or multipart form
	// when it exists, otherwise it returns an empty string `("")`.
	PostForm(key string) string

	// GetPostForm is like PostForm(key). It returns the specified key from a POST urlencoded
	// form or multipart form when it exists `(value, true)` (even when the value is an empty string),
	// otherwise it returns ("", false).
	// For example, during a PATCH request to update the user's email:
	//     email=mail@example.com  -->  ("mail@example.com", true) := GetPostForm("email") // set email to "mail@example.com"
	// 	   email=                  -->  ("", true) := GetPostForm("email") // set email to ""
	//                             -->  ("", false) := GetPostForm("email") // do nothing with email
	GetPostForm(key string) (string, bool)

	// PostFormArray returns a slice of strings for a given form key.
	// The length of the slice depends on the number of params with the given key.
	PostFormArray(key string) []string

	// GetPostFormArray returns a slice of strings for a given form key, plus
	// a boolean value whether at least one value exists for the given key.
	GetPostFormArray(key string) ([]string, bool)

	// PostFormMap returns a map for a given form key.
	PostFormMap(key string) map[string]string

	// GetPostFormMap returns a map for a given form key, plus a boolean value
	// whether at least one value exists for the given key.
	GetPostFormMap(key string) (map[string]string, bool)

	// FormFile returns the first file for the provided form key.
	FormFile(name string) (*multipart.FileHeader, error)

	// MultipartForm is the parsed multipart form, including file uploads.
	MultipartForm() (*multipart.Form, error)

	// SaveUploadedFile uploads the form file to specific dst.
	SaveUploadedFile(file *multipart.FileHeader, dst string) error

	// * REQUEST BODY
	// ******************************************************************

	// Bind automatically resolves binding objects according to content-Type
	// 	Content-Type                      | Binding      | Struct tag
	//  ----------------------------------|--------------|--------------------
	//  application/json                  | JSON binding | `json:"field_name"`
	//  application/xml                   | XML binding  | `xml:"field_name"`
	//  application/x-www-form-urlencoded | FORM binding | `form:"field_name"`
	//  multipart/form-data               | FORM binding | `form:"field_name"`
	//  GET request method                | FORM binding | `form:"field_name"`
	Bind(obj interface{}) error

	// BindJSON application/json
	BindJSON(obj interface{}) error

	// * RESPONSE RENDERING
	// ******************************************************************

	// SetCookie adds a Set-Cookie header to the ResponseWriter's headers.
	SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool)

	// Cookie return cookie value
	Cookie(name string) (string, error)

	// GetHeader returns value from request headers.
	// shortcut for c.Request.Header.Get(key)
	GetHeader(key string) string

	// Header is a intelligent shortcut for c.Writer.Header().Set(key, value).
	Header(key, value string)

	// Status sets the HTTP response code.
	Status(code int)
}
