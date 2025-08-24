package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	_ "github.com/pllus/main-fiber/docs"

	"main-fiber/big_workspace/internal/config"
	"main-fiber/big_workspace/internal/api/routes"
	"main-fiber/big_workspace/internal/database"
)

var client *mongo.Client

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Connect to the database
	client := database.ConnectMongo(cfg.MongoURI)
	defer client.Disconnect(nil)

	// Fiber app
	app := fiber.New()

	// Swagger API document for Faisu and Vincy
	app.Get("/docs/*", swagger.HandlerDefault)

	// Routes
	routes.SetupRoutes(app, client)

	// RUN SERVER
	log.Fetal(app.Listen(":" + cfg.Port))
}