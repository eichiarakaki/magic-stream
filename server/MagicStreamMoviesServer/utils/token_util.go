package utils

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/eichiarakaki/magic-stream/database"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var userCollection *mongo.Collection = database.OpenCollection("users")

// SignedDetails represents the custom JWT claims that will be embedded
// inside both the access token and the refresh token.
// It includes user identity information plus the standard registered claims.
type SignedDetails struct {
	Email                string `json:"email"`
	FirstName            string `json:"first_name"`
	LastName             string `json:"last_name"`
	Role                 string `json:"role"`
	UserID               string `json:"user_id"`
	jwt.RegisteredClaims        // Standard JWT fields (issuer, expiration, issuedAt…)
}

// Secret keys used to sign the access token and refresh token.
// They are loaded from environment variables.
var SecretKey = os.Getenv("SECRET_KEY")
var SecretRefreshKey = os.Getenv("SECRET_REFRESH_KEY")

// GenerateAllTokens creates and signs both an access token and a refresh token.
// The access token contains user information and expires in 1 hour.
func GenerateAllTokens(email, firstName, lastName, role, userID string) (string, string, error) {

	// Access token claims
	claims := &SignedDetails{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Role:      role,
		UserID:    userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "MagicStream",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)), // 1 hour
		},
	}

	// Create and sign the access token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		return "", "", err
	}

	// Refresh token claims
	refreshClaims := &SignedDetails{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Role:      role,
		UserID:    userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "MagicStream",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		},
	}

	// Create and sign the refresh token
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	signedRefreshToken, err := refreshToken.SignedString([]byte(SecretRefreshKey))
	if err != nil {
		return "", "", err
	}

	return signedToken, signedRefreshToken, nil
}

// UpdateAllTokens stores the new access token and refresh token in the user's MongoDB document.
// This is typically called after login or when refreshing tokens.
func UpdateAllTokens(token, refreshToken, userID string) (err error) {

	// Create a timeout context for the database update operation
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	// RFC3339 timestamp for updated_at
	updateAt, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

	// MongoDB update object
	updateData := bson.M{
		"$set": bson.M{
			"token":         token,
			"refresh_token": refreshToken,
			"updated_at":    updateAt,
		},
	}

	// Update the user document by ID
	_, err = userCollection.UpdateOne(ctx, bson.M{"user_id": userID}, updateData)
	if err != nil {
		return err
	}

	return nil
}

func GetAccessToken(c *gin.Context) (string, error) {
	authHeader := c.Request.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("no authorization header found")
	}

	tokenString := authHeader[len("Bearer "):]
	if tokenString == "" {
		return "", errors.New("no bearer token found")
	}

	return tokenString, nil
}

// ValidateToken validates a JWT string and returns its claims.
//
// How it works:
// 1. Parse the JWT and load its data into SignedDetails.
// 2. Verify the signature with the server's secret key.
// 3. Ensure the signing method is HMAC.
// 4. Check expiration.
// 5. Return claims or an error.
func ValidateToken(tokenString string) (*SignedDetails, error) {

	// Struct where we want to decode the claims
	claims := &SignedDetails{}

	// Parse and validate the token
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})
	if err != nil {
		return nil, err
	}

	// Ensure correct signing method (HS256)
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}

	// Check expiration timestamp
	if claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.New("token is expired")
	}

	return claims, nil
}

func GetUserIDFromContext(c *gin.Context) (string, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", errors.New("no user id found")
	}
	
	return userID.(string), nil
}
