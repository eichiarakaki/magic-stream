package database

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Here we do the requests of the content from the MongoDB database

// DBIsntance Returns a MongoDB instance
func DBInstance() *mongo.Client {
	// Here we're loading the environment variables
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	MongoURI := os.Getenv("MONGODB_URI")

	if MongoURI == "" {
		log.Fatal("MONGODB_URI environment variable not set")
	}

	//MongoDB := os.Getenv("MONGODB_DATABASE")
	//if MongoDB == "" {
	//	log.Fatal("MONGODB_DATABASE environment variable not set")
	//}

	// Connecting to the Mongo Database
	clientOptions := options.Client().ApplyURI(MongoURI)
	client, err := mongo.Connect(clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully connected to MongoDB!")
	//fmt.Println("MONGODB_DATABASE: ", MongoDB)
	fmt.Println("MONGODB_URI: ", MongoURI)

	return client
}

var Client *mongo.Client = DBInstance()

func OpenCollection(collectionName string) *mongo.Collection {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	databaseName := os.Getenv("DATABASE_NAME")
	if databaseName == "" {
		log.Fatal("DATABASE_NAME environment variable not set")
	}
	collection := Client.Database(databaseName).Collection(collectionName)

	if collection == nil {
		fmt.Println("Error fetching collection")
		return nil
	}
	return collection
}
