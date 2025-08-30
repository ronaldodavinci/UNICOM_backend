// @title Fiber MongoDB API
// @version 1.0
// @description This is a sample server for user management.
// @host localhost:8000
// @BasePath /

package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	_ "main-webbase/docs"

	"go.mongodb.org/mongo-driver/mongo"
	"main-webbase/config"
	"main-webbase/database"
	"main-webbase/internal/routes"
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
	routes.SetupRoutesUser(app, client)
	routes.SetupRoutesRole(app, client)
	routes.SetupRoutesUser_Role(app, client)

	// RUN SERVER
	log.Fatal(app.Listen(":" + cfg.Port))
}
