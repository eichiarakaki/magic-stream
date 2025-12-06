package controllers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/eichiarakaki/magic-stream/database"
	"github.com/eichiarakaki/magic-stream/models"
	"github.com/eichiarakaki/magic-stream/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"google.golang.org/genai"
)

var validate = validator.New()

func GetMovies(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 100*time.Second)
		defer cancel()

		var movies []models.Movie

		// Doing a request to the MongoDB with NO filters.
		cursor, err := database.OpenCollection("movies", client).Find(ctx, bson.M{})
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

func GetMovie(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 100*time.Second)
		defer cancel()

		movieID := c.Param("imdb_id") // You can obtain the parameter by using :imdb_id when mapping the URL
		if movieID == "" {
			c.JSON(http.StatusNotFound, gin.H{"Error": "Movie ID not found"})
		}

		var movie models.Movie
		// request the specific video by filtering by imdb_id
		err := database.OpenCollection("movies", client).FindOne(ctx, bson.M{"imdb_id": movieID}).Decode(&movie)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Movie not found"})
		}

		c.JSON(http.StatusOK, movie)
	}
}

func AddMovie(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 100*time.Second)
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
		result, err := database.OpenCollection("movies", client).InsertOne(ctx, movie)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to add movie"})
			return
		}

		c.JSON(http.StatusCreated, result)
	}
}

// AdminReviewUpdate gets a body containing admin_review which later is sent to a LLM with custom prompts
// then the results are updated to the specified video/movie.
func AdminReviewUpdate(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 100*time.Second)
		defer cancel()

		role, err := utils.GetUserRoleFromContext(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "user role not found"})
			return
		}
		log.Println(role)
		if role != "ADMIN" {
			c.JSON(http.StatusForbidden, gin.H{"Error": "You do not have admin role to update"})
			return
		}

		movieID := c.Param("imdb_id")
		if movieID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Movie ID required"})
			return
		}

		var req struct {
			AdminReview string `json:"admin_review"`
		}
		var resp struct {
			RankingName string `json:"ranking_name"`
			AdminReview string `json:"admin_review"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}

		sentiment, rankVal, err := GetReviewRanking(req.AdminReview, client, c)
		if err != nil {
			c.JSON(
				http.StatusInternalServerError,
				gin.H{"error": "Error getting review ranking",
					"details": err.Error()},
			)
			return
		}

		filter := bson.M{"imdb_id": movieID}
		update := bson.M{"$set": bson.M{
			"admin_review": req.AdminReview,
			"ranking": bson.M{
				"ranking_name":  sentiment,
				"ranking_value": rankVal,
			},
		}}

		result, err := database.OpenCollection("movies", client).UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to update review"})
			return
		}

		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"Error": "Movie not found"})
			return
		}

		resp.RankingName = sentiment
		resp.AdminReview = req.AdminReview

		c.JSON(http.StatusOK, resp)
	}
}

func GetReviewRanking(admin_review string, client *mongo.Client, c *gin.Context) (string, int, error) {
	rankings, err := GetRankings(client, c)
	if err != nil {
		return "", 0, err
	}

	sentimentDelimited := ""

	for _, ranking := range rankings {
		if ranking.RankingValue != 999 && ranking.RankingValue != 0 {
			sentimentDelimited = sentimentDelimited + ranking.RankingName + ","
		}
	}

	sentimentDelimited = strings.Trim(sentimentDelimited, ",")

	err = godotenv.Load(".env")
	if err != nil {
		log.Println("Warning: .env file not found")
		// return "", 0, err
	}

	base_prompt_template := os.Getenv("BASE_PROMPT_TEMPLATE")
	base_prompt := strings.Replace(base_prompt_template, "{rankings}", sentimentDelimited, -1)

	// Connecting to Gemini
	ctx := context.Background()
	// The client gets the API key from the environment variable `GEMINI_API_KEY`.
	genai_client, err := genai.NewClient(ctx, nil)
	if err != nil {
		log.Println("Warning: genai client failed", err)
		return "", 0, err
	}
	response, err := genai_client.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash",
		genai.Text(base_prompt+admin_review),
		nil,
	)
	if err != nil {
		return "", 0, err
	}

	rankVal := 0
	for _, ranking := range rankings {
		if ranking.RankingName == response.Text() {
			rankVal = ranking.RankingValue
			break
		}
	}

	return response.Text(), rankVal, nil
}

// GetRankings request the existing rankings data from the MongoDB
func GetRankings(client *mongo.Client, c *gin.Context) ([]models.Ranking, error) {
	var rankings []models.Ranking

	var ctx, cancel = context.WithTimeout(c.Request.Context(), 100*time.Second)
	defer cancel()

	cursor, err := database.OpenCollection("rankings", client).Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			log.Println(err)
		}
	}(cursor, ctx)

	if err = cursor.All(ctx, &rankings); err != nil {
		return nil, err
	}

	return rankings, nil
}

// GetRecommendedMovies returns a handler that recommends movies
// based on a user's favorite genres.
//
// HOW THIS ENDPOINT WORKS
// -------------------------------------------------------------
// 1. Extract the user ID from the request context (Auth middleware).
// 2. Get a list of the user’s favorite genres from MongoDB.
// 3. Load environment variable RECOMMENDED_MOVIE_LIMIT (default = 5).
// 4. Query the movies collection:
//   - Filter: any movie whose genre matches user's favorites
//   - Sort: by ranking.ranking_value (ascending → better ranking first)
//   - Limit: maximum number of movies to return
//
// 5. Decode results into []models.Movie
// 6. Return JSON with recommended movies.
//
// MONGO FILTER EXPLAINED
// -------------------------------------------------------------
//
//	filter := bson.M{
//	    "genre.genre_name": bson.M{"$in": favorite_genres}
//	}
//
// "$in" means: select all movies where genre.genre_name appears
// in the user's favorite genres list.
//
// MONGO FIND OPTIONS
// -------------------------------------------------------------
// SetSort → order results by a specific field
// SetLimit → maximum number of documents to return
//
// EXPECTED MOVIE DOCUMENT STRUCTURE
// -------------------------------------------------------------
//
//	{
//	  "title": "Inception",
//	  "genre": [
//	      { "genre_name": "Sci-Fi" },
//	      { "genre_name": "Action" }
//	  ],
//	  "ranking": {
//	      "ranking_value": 120
//	  }
//	}
func GetRecommendedMovies(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {

		// 1. Extract user ID from context (set by AuthMiddleware)
		userID, err := utils.GetUserIDFromContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"Error": "Unauthorized"})
			return // ← IMPORTANT! Stop execution
		}

		// 2. Fetch user's favorite genres from DB
		favorite_genres, err := GetUsersFavoriteGenres(userID, client, c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}

		// 3. Load .env (not required in production, but fine for local dev)
		err = godotenv.Load(".env")
		if err != nil {
			log.Println("Warning: .env file not found")
		}

		// Default recommended limit = 5
		var recommendedMovieLimitVal int64 = 5

		// If the env var exists, convert it to integer
		recommendedMovieLimitStr := os.Getenv("RECOMMENDED_MOVIE_LIMIT")
		if recommendedMovieLimitStr != "" {
			recommendedMovieLimitVal, _ = strconv.ParseInt(recommendedMovieLimitStr, 10, 64)
		}

		// 4. Mongo query options: sort + limit
		findOptions := options.Find()
		findOptions.SetSort(bson.D{{Key: "ranking.ranking_value", Value: 1}}) // ascending
		findOptions.SetLimit(recommendedMovieLimitVal)

		// 5. Filter: match movies whose genre_name is in the user’s favorites
		filter := bson.M{
			"genre.genre_name": bson.M{
				"$in": favorite_genres,
			},
		}

		// 6. Create context with timeout
		ctx, cancel := context.WithTimeout(c.Request.Context(), 100*time.Second)
		defer cancel()

		// 7. Perform the query
		cursor, err := database.OpenCollection("movies", client).Find(ctx, filter, findOptions)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}

		// Ensure the cursor is closed after usage
		defer func(cursor *mongo.Cursor, ctx context.Context) {
			if err := cursor.Close(ctx); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			}
		}(cursor, ctx)

		// 8. Decode results into Go struct
		var recommendedMovies []models.Movie
		if err = cursor.All(ctx, &recommendedMovies); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Error fetching recommended movies"})
			return
		}

		// 9. Return movie list as JSON
		c.JSON(http.StatusOK, recommendedMovies)
	}
}

// GetUsersFavoriteGenres retrieves a list of genre names from the user's
// "favourite_genres" array stored in MongoDB.
//
// HOW IT WORKS:
// ------------------------------------------------------
// 1. Creates a timeout context so the DB call does not hang forever.
// 2. Uses FindOne() to retrieve ONLY ONE user document that matches the filter.
// 3. Uses SetProjection() to tell MongoDB which fields we want to return.
//   - In this case, ONLY "favourite_genres.genre_name" and not the whole doc.
//
// 4. Decodes the MongoDB result into a generic bson.M (map).
// 5. Extracts the list of genres from the nested BSON array.
// 6. Converts it into a []string of genre names.
// 7. Returns the list of names.
//
// IMPORTANT MONGO TERMS:
// ------------------------------------------------------
//
//   - FindOne(filter, options):
//     Searches for ONE document matching 'filter'.
//
//   - SetProjection(projection):
//     Selects ONLY specific fields from the document.
//     Example: {"favourite_genres.genre_name": 1} → include this nested field.
//     "_id": 0 → exclude MongoDB's ID.
//
// • bson.M: a MongoDB document represented as a map[string]interface{}
//
// • bson.A: a MongoDB array represented as []interface{}
//
// DOCUMENT STRUCTURE EXPECTED FROM MONGO:
// ------------------------------------------------------
// user document example:
//
//	{
//	  "user_id": "123",
//	  "favourite_genres": [
//	     { "genre_id": 1, "genre_name": "Action" },
//	     { "genre_id": 2, "genre_name": "Horror" }
//	  ]
//	}
//
// The projection returns ONLY this:
//
//	{
//	  "favourite_genres": [
//	     { "genre_name": "Action" },
//	     { "genre_name": "Horror" }
//	  ]
//	}
func GetUsersFavoriteGenres(userID string, client *mongo.Client, c *gin.Context) ([]string, error) {
	// Create a 100s timeout so Mongo doesn't block forever.
	ctx, cancel := context.WithTimeout(c.Request.Context(), 100*time.Second)
	defer cancel()

	// Filter → find the document where user_id == given userID
	filter := bson.M{"user_id": userID}

	// Projection → tell Mongo which fields to return.
	// "1" means include, "0" means exclude.
	projection := bson.M{
		"favourite_genres.genre_name": 1, // include genre names inside the array
		"_id":                         0, // exclude Mongo’s default ID
	}

	// Apply projection to the query
	opts := options.FindOne().SetProjection(projection)

	// We'll decode the result into a generic map
	var result bson.M

	// Actually run the query
	err := database.OpenCollection("users", client).FindOne(ctx, filter, opts).Decode(&result)
	if err != nil {
		// If no user found → return empty list instead of an error
		if err == mongo.ErrNoDocuments {
			return []string{}, nil
		}
		return nil, err
	}

	// Extract the array: result["favourite_genres"] must be an array (bson.A)
	favGenresArray, ok := result["favourite_genres"].(bson.A)
	if !ok {
		return []string{}, errors.New("unable to retrieve favorite genres for user")
	}

	// Convert BSON array → Go []string
	var genreNames []string
	for _, genre := range favGenresArray {
		// Each element is a BSON document e.g. { "genre_name": "Action" }
		if genreMap, ok := genre.(bson.D); ok {
			// bson.D is a slice of key-value pairs
			for _, elem := range genreMap {
				if elem.Key == "genre_name" {
					if name, ok := elem.Value.(string); ok {
						genreNames = append(genreNames, name)
					}
				}
			}
		}
	}

	return genreNames, nil
}

func GetGenres(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(c.Request.Context(), 100*time.Second)
		defer cancel()

		var genreCollection *mongo.Collection = database.OpenCollection("genres", client)

		cursor, err := genreCollection.Find(ctx, bson.D{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching movie genres", "details": err.Error()})
			return
		}
		defer func(cursor *mongo.Cursor, ctx context.Context) {
			err := cursor.Close(ctx)
			if err != nil {
				log.Println(err)
			}
		}(cursor, ctx)

		var genres []models.Genre
		if err = cursor.All(ctx, &genres); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, genres)
	}
}
