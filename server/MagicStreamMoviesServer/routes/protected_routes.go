package routes

import (
	controller "github.com/eichiarakaki/magic-stream/controllers"
	"github.com/eichiarakaki/magic-stream/middleware"
	"github.com/gin-gonic/gin"
)

func SetupProtectedRoutes(router *gin.Engine) {
	router.Use(middleware.AuthMiddleware())

	router.GET("/movie/:imdb_id", controller.GetMovie())
	router.POST("/add-movie", controller.AddMovie())
	router.GET("/recommended-movies", controller.GetRecommendedMovies())
}
