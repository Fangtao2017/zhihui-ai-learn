package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	JWTSecretKey string
	MongoDBURI   string
	DBName       string
)

func Init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	JWTSecretKey = os.Getenv("JWT_SECRET_KEY")
	if JWTSecretKey == "" {
		log.Fatal("JWT_SECRET_KEY is not set in .env file")
	}

	MongoDBURI = os.Getenv("MONGODB_URI")
	if MongoDBURI == "" {
		MongoDBURI = "mongodb://localhost:27017" // 默认值
	}

	DBName = os.Getenv("DB_NAME")
	if DBName == "" {
		DBName = "admin" // 默认值
	}
}
