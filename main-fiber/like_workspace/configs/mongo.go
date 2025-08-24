package configs

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

func ConnectMongo() *mongo.Client {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().
		ApplyURI("mongodb+srv://nitisarath:p5U424lf6bmlFCBw@cluster0.m01vueg.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0").
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
