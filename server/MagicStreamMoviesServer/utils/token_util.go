package utils

import (
	"context"
	"os"
	"time"

	"github.com/eichiarakaki/magic-stream/database"
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
