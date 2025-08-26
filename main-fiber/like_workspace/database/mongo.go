package configs

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

var PostCollection *mongo.Collection

func ConnectMongo() *mongo.Client {
	// โหลด .env
	if err := godotenv.Load(); err != nil {
		log.Fatal("❌ .env file not found")
	}

	uri := os.Getenv("MONGO_URI")
	dbName := os.Getenv("DB_NAME")

	if uri == "" || dbName == "" {
		log.Fatal("❌ MONGO_URI or DB_NAME not set in environment")
	}

	opts := options.Client().
		ApplyURI(uri).
		SetServerAPIOptions(options.ServerAPI(options.ServerAPIVersion1))

	client, err := mongo.Connect(opts)
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}

	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal("Failed to ping MongoDB:", err)
	}

	fmt.Println("✅ Connected to MongoDB!")
	PostCollection = client.Database(dbName).Collection("posts")

	return client
}

func DisconnectMongo(client *mongo.Client) {
	if err := client.Disconnect(context.TODO()); err != nil {
		panic(err)
	}
}
