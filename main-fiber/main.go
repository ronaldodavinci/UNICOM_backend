package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	_ "github.com/pllus/main-fiber/docs"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

// Make the client a global variable so other functions can access it
var client *mongo.Client

func main() {
	// --- MongoDB Connection Setup ---
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI("mongodb+srv://nitisarath:p5U424lf6bmlFCBw@cluster0.m01vueg.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0").SetServerAPIOptions(serverAPI)

	var err error
	client, err = mongo.Connect(opts)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal("Failed to ping MongoDB:", err)
	}
	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")

	// --- Fiber App Setup ---
	app := fiber.New()

	app.Get("/docs/*", swagger.HandlerDefault)

	app.Get("/hello", getHello)

	// Add the new GET route for fetching user data
	app.Get("/Post", getUserHandler)

	// privateRoutes := app.Group("/", middleware.AuthMiddleware)
	// app.Get("/User", example.GetDataHandler)
	// app.Get("/Post", example.GetDataHandler_post)
	// privateRoutes.Get("/Protected", example.GetDataHandler)
	// privateRoutes.Get("/Protected", example.GetDataHandler)

	//server variables
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Fatal(app.Listen(":" + port))
}

func getHello(c *fiber.Ctx) error {
	return c.SendString("Hello, World!")
}

func getUserHandler(c *fiber.Ctx) error {
	db := client.Database("User_1")
	collection := db.Collection("User")

	// Find all documents in the collection
	cursor, err := collection.Find(context.TODO(), bson.D{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to query database",
			"data":    nil,
		})
	}
	defer cursor.Close(context.TODO())

	var users []bson.M
	if err = cursor.All(context.TODO(), &users); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to decode documents",
			"data":    nil,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Users fetched successfully",
		"data":    users,
	})
}
