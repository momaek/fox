package main

import "fox/engine"

func main() {
	app := engine.New()

	app.GET("/", func(c *engine.Context) {
		c.Send("Hello, World 👋!")
	})

	app.Listen(3000)
}
