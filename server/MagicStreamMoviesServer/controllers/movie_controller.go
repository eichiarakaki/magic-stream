package controllers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/eichiarakaki/magic-stream/database"
	"github.com/eichiarakaki/magic-stream/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var movieCollection *mongo.Collection = database.OpenCollection("movies")

func GetMovies() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var movies []models.Movie

		// Doing a request to the MongoDB with NO filters.
		cursor, err := movieCollection.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to fetch movies"})
		}
		// Closing the final cursor to prevent memory leaks.
		defer func(cursor *mongo.Cursor, ctx context.Context) {
			err := cursor.Close(ctx)
			if err != nil {
				fmt.Println(err)
			}
		}(cursor, ctx)

		// Converting the cursor to the movies slice
		if err = cursor.All(ctx, &movies); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to decode movies"})
		}

		c.JSON(http.StatusOK, gin.H{"movies": movies})
	}
}
