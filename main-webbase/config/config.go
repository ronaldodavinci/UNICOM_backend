package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoURI string
	MongoDB  string
	Port     string
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func LoadConfig() Config {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	cfg := Config{
		MongoURI: getEnv("MONGO_URI", "mongodf://localhost:27017"),
		MongoDB:  getEnv("MONGO_DB", "creatorDatabase"),
		Port:	 getEnv("PORT", "3000"),
	}
	return cfg
}