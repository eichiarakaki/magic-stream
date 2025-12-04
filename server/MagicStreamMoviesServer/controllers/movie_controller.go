package controllers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/eichiarakaki/magic-stream/database"
	"github.com/eichiarakaki/magic-stream/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Gets all the movies
var movieCollection *mongo.Collection = database.OpenCollection("movies")

var validate = validator.New()

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

		c.JSON(http.StatusOK, movies)
	}
}

func GetMovie() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		movieID := c.Param("imdb_id") // You can obtain the parameter by using :imdb_id when mapping the URL
		if movieID == "" {
			c.JSON(http.StatusNotFound, gin.H{"Error": "Movie ID not found"})
		}

		var movie models.Movie
		// request the specific video by filtering by imdb_id
		err := movieCollection.FindOne(ctx, bson.M{"imdb_id": movieID}).Decode(&movie)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Movie not found"})
		}

		c.JSON(http.StatusOK, movie)
	}
}

func AddMovie() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var movie models.Movie

		err := c.ShouldBindJSON(&movie)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid movie data"})
			return
		}

		if err := validate.Struct(&movie); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Validation failed", "details": err.Error()})
			return
		}

		// Inserting a new movie to the MongoDB
		result, err := movieCollection.InsertOne(ctx, movie)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to add movie"})
			return
		}

		c.JSON(http.StatusCreated, result)
	}
}
