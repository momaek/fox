package main

import "fox/engine"

func main() {

	router := engine.Default()

	router.GET("/ping", func(c *engine.Context) (resp interface{}, err error) {
		resp = engine.H{
			"message": "pong",
		}
		return
	})

	router.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
