package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
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
        MongoURI: getEnv("MONGO_URI", "mongodb://localhost:27017"),
        Port:	 getEnv("PORT", "3000"),
    }
    return cfg
}

func ConnectMongo() *mongo.Client {
	cfg := LoadConfig()
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().
		ApplyURI(cfg.MongoURI).
		SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(opts)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal("Failed to ping MongoDB:", err)
	}
	fmt.Println("✅ Pinged your deployment. Connected to MongoDB!")

	return client
}

func DisconnectMongo(client *mongo.Client) {
	if err := client.Disconnect(context.TODO()); err != nil {
		panic(err)
	}
}
