package db

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	DB *mongo.Database
)

const (
	UserCollection    = "users"
	ChatCollection    = "chats"
	MessageCollection = "messages"
)

// InitDB initializes the database connection
func InitDB() error {
	// Get configuration from environment variables
	mongoURI := os.Getenv("MONGODB_URI")
	dbName := os.Getenv("DB_NAME")

	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}
	if dbName == "" {
		dbName = "admin"
	}

	// Set connection timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create client connection options
	clientOptions := options.Client().ApplyURI(mongoURI)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Printf("Failed to connect to MongoDB: %v", err)
		return err
	}

	// Test connection
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Printf("Failed to ping MongoDB: %v", err)
		return err
	}

	// Initialize global DB variable
	DB = client.Database(dbName)
	log.Printf("Successfully connected to MongoDB database: %s", dbName)

	return nil
}

// GetCollection gets the specified collection
func GetCollection(name string) *mongo.Collection {
	if DB == nil {
		log.Fatal("Database connection not initialized")
	}
	return DB.Collection(name)
}

// CloseDB closes the database connection
func CloseDB() {
	if DB != nil {
		if client := DB.Client(); client != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := client.Disconnect(ctx); err != nil {
				log.Printf("Error disconnecting from MongoDB: %v", err)
			}
		}
	}
}
