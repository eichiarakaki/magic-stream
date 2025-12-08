package routes

import (
	controller "github.com/eichiarakaki/magic-stream/controllers"
	"github.com/eichiarakaki/magic-stream/middleware"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func SetupProtectedRoutes(router *gin.Engine, client *mongo.Client) {
	router.Use(middleware.AuthMiddleware())

	router.GET("/movie/:imdb_id", controller.GetMovie(client))
	router.POST("/add-movie", controller.AddMovie(client))
	router.GET("/recommended-movies", controller.GetRecommendedMovies(client))
	router.PATCH("/update-review/:imdb_id", controller.AdminReviewUpdate(client))
	router.POST("/logout", controller.LogoutUser(client))
}
