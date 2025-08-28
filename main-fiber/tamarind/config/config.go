package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoURI string
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
		MongoURI: getEnv("MONGO_URI", "mongodb+srv://root:971397@cluster01.wawl1f9.mongodb.net/"),
		Port:	 getEnv("PORT", "3000"),
	}
	return cfg
}