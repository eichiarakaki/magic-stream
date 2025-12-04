package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/eichiarakaki/magic-stream/database"
	"github.com/eichiarakaki/magic-stream/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection("users")

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// RegisterUser handles the registration of a new user.
// It validates input data, checks for duplicate email addresses,
// hashes the password, and inserts the user into the MongoDB collection.
func RegisterUser() gin.HandlerFunc {
	return func(c *gin.Context) {

		// Parse and bind the incoming JSON payload to the User model
		var user models.User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"Error":   "Invalid input data",
				"details": err.Error(),
			})
			return
		}

		// Validate the fields of the User struct using the validator library
		validate := validator.New()
		if err := validate.Struct(user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"Error":   "Validation failed",
				"details": err.Error(),
			})
			return
		}

		// Hash the user's password before storing it
		hashedPassword, err := HashPassword(user.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"Error":   "Failed to hash password",
				"details": err.Error(),
			})
			return
		}

		// Create a context with a 100-second timeout for database operations
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// Ensure the email is unique by counting documents with the same email
		count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"Error":   "Failed to check existing user",
				"details": err.Error(),
			})
			return
		}
		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{
				"Error":   "User already exists",
				"details": "User already exists",
			})
			return
		}

		// Prepare the user object for database insertion
		user.UserID = bson.NewObjectID().Hex()
		user.Password = hashedPassword
		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()

		// Insert the user into the MongoDB collection
		result, err := userCollection.InsertOne(ctx, user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"Error":   "Failed to add user",
				"details": err.Error(),
			})
			return
		}

		// Return the insertion result to the client
		c.JSON(http.StatusCreated, result)
	}
}
