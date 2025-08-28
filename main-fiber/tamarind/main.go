package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"

	"tamarind/internal/routes"
	"tamarind/internal/database"
)

func main() {
	// Connect to MongoDB
	client := database.ConnectMongo()
	defer database.DisconnectMongo(client)

	// Fiber app
	app := fiber.New() // ceate a new Fiber app
	app.Get("/docs/*", swagger.HandlerDefault) // Enable swagger API

	// Register routes
	routes.Register(app, client)

	// Server variables
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000" // default port
	}

	log.Fatal(app.Listen(":" + port)) // start
}
