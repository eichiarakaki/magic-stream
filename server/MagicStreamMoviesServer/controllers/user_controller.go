package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/eichiarakaki/magic-stream/database"
	"github.com/eichiarakaki/magic-stream/models"
	"github.com/eichiarakaki/magic-stream/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/bcrypt"
)

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
func RegisterUser(client *mongo.Client) gin.HandlerFunc {
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
		var ctx, cancel = context.WithTimeout(c.Request.Context(), 100*time.Second)
		defer cancel()

		// Ensure the email is unique by counting documents with the same email
		count, err := database.OpenCollection("users", client).CountDocuments(ctx, bson.M{"email": user.Email})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"Error":   "Failed to check existing user",
				"details": err.Error(),
			})
			return
		}
		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{
				"Error": "User already exists",
			})
			return
		}

		// Prepare the user object for database insertion
		user.UserID = bson.NewObjectID().Hex()
		user.Password = hashedPassword
		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()

		// Insert the user into the MongoDB collection
		result, err := database.OpenCollection("users", client).InsertOne(ctx, user)
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

// LoginUser handles the login process for a user.
// It validates the incoming credentials, checks whether the user exists,
// compares the provided password with the stored hashed password,
// generates new JWT access and refresh tokens, updates them in the database,
// and finally returns user information along with the tokens.
func LoginUser(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Parse JSON payload into the UserLogin model
		var userLogin models.UserLogin
		if err := c.ShouldBindJSON(&userLogin); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"Error": "Invalid input data",
			})
			return
		}

		// Create a context for database operations
		var ctx, cancel = context.WithTimeout(c.Request.Context(), 100*time.Second)
		defer cancel()

		// Try to find the user in the database by email
		var foundUser models.User
		err := database.OpenCollection("users", client).FindOne(ctx, bson.M{"email": userLogin.Email}).Decode(&foundUser)
		if err != nil {
			// User was not found, or error occurred
			c.JSON(http.StatusUnauthorized, gin.H{
				"Error": "Invalid email",
			})
			return
		}

		// Compare the provided password with the stored hashed password
		err = bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(userLogin.Password))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"Error": "Invalid password",
			})
			return
		}

		// Generate access and refresh tokens for the authenticated user
		token, refreshToken, err := utils.GenerateAllTokens(
			foundUser.Email,
			foundUser.FirstName,
			foundUser.LastName,
			foundUser.Role,
			foundUser.UserID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"Error": "Failed to generate the tokens",
			})
			return
		}

		// Store the new tokens in the user's document in MongoDB
		err = utils.UpdateAllTokens(token, refreshToken, foundUser.UserID, client, c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"Error": "Failed to update the tokens",
			})
			return
		}

		http.SetCookie(c.Writer, &http.Cookie{
			Name:  "access_token",
			Value: token,
			Path:  "/",
			// Domain:   "localhost",
			MaxAge:   86400,
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteNoneMode,
		})
		http.SetCookie(c.Writer, &http.Cookie{
			Name:  "refresh_token",
			Value: refreshToken,
			Path:  "/",
			// Domain:   "localhost",
			MaxAge:   604800,
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteNoneMode,
		})

		// Return user information and the generated tokens
		c.JSON(http.StatusOK, models.UserResponse{
			UserID:         foundUser.UserID,
			FirstName:      foundUser.FirstName,
			LastName:       foundUser.LastName,
			Email:          foundUser.Email,
			Role:           foundUser.Role,
			FavoriteGenres: foundUser.FavoriteGenres,
			Token:          token,
			RefreshToken:   refreshToken,
		})
	}
}

// LogoutUser logs out ONLY the currently authenticated user.
func LogoutUser(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDValue, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not logged in"})
			return
		}
		userID, ok := userIDValue.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert user ID to string"})
			return
		}

		// Remove tokens in database
		// Passing empty strings "" means: no valid tokens stored
		err := utils.UpdateAllTokens("", "", userID, client, c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Error logging out",
				"details": err.Error(),
			})
			return
		}

		// Delete the access token cookie in the browser.
		// MaxAge: -1 tells the browser to immediately remove the cookie.
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     "access_token",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,   // expire immediately
			Secure:   true, // send only over HTTPS
			HttpOnly: true, // JS cannot read the cookie
			SameSite: http.SameSiteStrictMode,
		})

		// Final response to the client
		c.JSON(http.StatusOK, gin.H{
			"message": "Successfully logged out",
		})
	}
}

func RefreshTokenHandler(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(c.Request.Context(), 100*time.Second)
		defer cancel()

		refreshToken, err := c.Cookie("refresh_token")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to retrieve refresh token from cookie"})
			return
		}

		claim, err := utils.ValidateRefreshToken(refreshToken)
		if err != nil || claim == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
			return
		}

		var userCollection *mongo.Collection = database.OpenCollection("users", client)

		var user models.User
		err = userCollection.FindOne(ctx, bson.D{{Key: "user_id", Value: claim.UserID}}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}

		newToken, newRefreshToken, _ := utils.GenerateAllTokens(user.Email, user.FirstName, user.LastName, user.Role, user.UserID)
		err = utils.UpdateAllTokens(newToken, newRefreshToken, user.UserID, client, c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update the tokens"})
			return
		}

		c.SetCookie("access_token", newToken, 86400, "/", "localhost", true, true)
		c.SetCookie("refresh_token", newRefreshToken, 604800, "/", "localhost", true, true)

		c.JSON(http.StatusOK, gin.H{"message": "Successfully refreshed the tokens"})
	}
}
