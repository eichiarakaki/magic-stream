package middleware

import (
	"net/http"

	"github.com/eichiarakaki/magic-stream/utils"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware is a Gin middleware that protects routes by requiring a valid JWT.
// It performs 3 steps:
//
//  1. Extract the access token from the Authorization header.
//     Expected format: "Authorization: Bearer <access_token>"
//
// 2. Validate the token using utils.ValidateToken(tokenString).
//
//   - Verifies the signature
//
//   - Checks token expiration
//
//   - Extracts the user claims (userID, role, email, etc.)
//
//     3. Stores the claims inside Gin's context so that other handlers
//     can access the authenticated user's information.
//
// If anything fails → the request is rejected with HTTP 401 and aborted.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		// 1. Extract JWT from Authorization header
		// utils.GetAccessToken should return only the raw token string
		tokenString, err := utils.GetAccessToken(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"Error": err.Error()})
			c.Abort()
			return
		}
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"Error": "token is empty"})
			c.Abort()
			return
		}

		// 2. Validate and parse JWT claims
		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"Error": err.Error()})
			c.Abort()
			return
		}

		// 3. Store user data inside Gin context
		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)

		c.Next()
	}
}
