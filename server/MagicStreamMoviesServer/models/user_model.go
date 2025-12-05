package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type User struct {
	ID             bson.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	UserID         string        `bson:"user_id" json:"user_id"`
	FirstName      string        `bson:"first_name" json:"first_name" validate:"required,min=2,max=100"`
	LastName       string        `bson:"last_name" json:"last_name" validate:"required"`
	Email          string        `bson:"email" json:"email" validate:"required"`
	Password       string        `bson:"password" json:"password" validate:"required,min=6"`
	Role           string        `bson:"role" json:"role" validate:"oneof=ADMIN USER"`
	CreatedAt      time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time     `bson:"updated_at" json:"updated_at"`
	Token          string        `bson:"token" json:"token"` // Used for authentication and authorization
	RefreshToken   string        `bson:"refresh_token" json:"refresh_token"`
	FavoriteGenres []Genre       `bson:"favourite_genres" json:"favourite_genres" validate:"required,min=1,dive"`
}
