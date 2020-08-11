package main

import "fox/engine"

func main() {

	router := engine.Default()

	router.GET("/ping", func(c *engine.Context) {
		c.JSON(200, engine.H{
			"message": "pong",
		})
	})

	router.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
