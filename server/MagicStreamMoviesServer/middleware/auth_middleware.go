package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/eichiarakaki/magic-stream/database"
	"github.com/eichiarakaki/magic-stream/models"
	"github.com/eichiarakaki/magic-stream/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// AuthMiddleware validates the JWT signature/expiry and also verifies that the
// presented access token matches the one stored in the database for the user.
// This allows immediate revocation (logout) by clearing the token in DB.
func AuthMiddleware(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1) Extract token string from header/cookie
		tokenString, err := utils.GetAccessToken(c)
		if err != nil || tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid access token"})
			c.Abort()
			return
		}

		// 2) Validate signature and expiry, obtain claims
		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token", "details": err.Error()})
			c.Abort()
			return
		}

		// 3) Query DB to ensure the token matches the current stored token for this user
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var dbUser models.User
		// Adjust the filter according to how you store the user identifier: here we use user_id field
		filter := bson.M{"user_id": claims.UserID}

		err = database.OpenCollection("users", client).FindOne(ctx, filter).Decode(&dbUser)
		if err != nil {
			// Not found or DB error → unauthorized
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found or db error", "details": err.Error()})
			c.Abort()
			return
		}

		// 4) Compare presented token with stored token
		// If you store hashed tokens, compare the hash instead.
		if dbUser.Token == "" || dbUser.Token != tokenString {
			// Token was revoked/rotated or does not match → unauthorized
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token is not valid (revoked or rotated)"})
			return
		}

		// 5) Put relevant info into context and continue
		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}
