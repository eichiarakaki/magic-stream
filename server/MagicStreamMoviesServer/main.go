package main

import (
	controller "github.com/eichiarakaki/magic-stream/controllers"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	router.GET("/hello", func(c *gin.Context) { // Creating a test EndPoint
		c.String(200, "Hello World")
	})

	router.GET("/movies", controller.GetMovies())

	if err := router.Run(":8080"); err != nil {
		panic(err)
	}
}
