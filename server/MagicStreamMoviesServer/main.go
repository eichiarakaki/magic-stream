package main

import (
	"github.com/eichiarakaki/magic-stream/database"
	"github.com/eichiarakaki/magic-stream/routes"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func main() {
	router := gin.Default()

	var client *mongo.Client = database.Connect()

	routes.SetupUnprotectedRoutes(router, client)
	routes.SetupProtectedRoutes(router, client)

	if err := router.Run(":8080"); err != nil {
		panic(err)
	}
}
