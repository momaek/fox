package engine

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRouterHandle(t *testing.T) {

	Convey("Register all HTTP methods routes", t, func() {

		var router = new(Router)

		Convey("unknown HTTP method", func() {
			So(func() {
				router.Handle("", "foo")
			}, ShouldPanic)
		})

		Convey("Nil handler", func() {
			So(func() {
				router.GET("foo")
			}, ShouldPanic)
		})

		Convey("Regist handle", func() {
			So(func() {
				router.GET("", func() string { return "ROOT" })
			}, ShouldNotPanic)

			So(func() {
				router.Any("any", func() string { return "Any" })
			}, ShouldNotPanic)

			So(func() {
				router.GET("get", func() string { return "GET" })
			}, ShouldNotPanic)

			So(func() {
				router.POST("post", func() string { return "POST" })
			}, ShouldNotPanic)

			So(func() {
				router.PUT("put", func() string { return "PUT" })
			}, ShouldNotPanic)

			So(func() {
				router.PATCH("patch", func() string { return "PATCH" })
			}, ShouldNotPanic)

			So(func() {
				router.DELETE("delete", func() string { return "DELETE" })
			}, ShouldNotPanic)

			So(func() {
				router.HEAD("head", func() string { return "HEAD" })
			}, ShouldNotPanic)

			So(func() {
				router.OPTIONS("options", func() string { return "OPTIONS" })
			}, ShouldNotPanic)
		})

		Convey("Multiple registrations", func() {

			var router = new(Router)

			So(func() {
				router.Any("any", func() string { return "Any" })
				router.Any("any", func() string { return "Any" })
			}, ShouldPanic)

			So(func() {
				router.Any("any", func() string { return "Any" })
				router.Any("/any", func() string { return "Any" })
			}, ShouldPanic)

			So(func() {
				router.GET("get", func() string { return "GET" })
				router.GET("/get", func() string { return "GET" })
			}, ShouldPanic)
		})
	})

}

func TestRouterGroup(t *testing.T) {

	Convey("Register all HTTP methods routes", t, func() {

		var router = new(Router)

		var comments = func() string {
			return "projects/:id/comments"
		}

		router.Group("projects/:id", func(group *Router) {
			group.GET("comments", comments)
		})

	})

}
