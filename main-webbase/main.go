// @title Fiber MongoDB API
// @version 1.0
// @description This is a sample server for user management.
// @host localhost:8000
// @BasePath /

package main

import (
	"log"

	_ "main-webbase/docs"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"

	"main-webbase/config"
	"main-webbase/database"
	"main-webbase/internal/routes"

	"go.mongodb.org/mongo-driver/mongo"
)

var client *mongo.Client

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Connect to the database
	client := database.ConnectMongo(cfg.MongoURI, cfg.MongoDB)
	defer client.Disconnect(nil)

	// Fiber app
	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", "68bd6ff6f80438824239b8a9") // 👈 ใช้ ObjectID จริงของ user
		c.Locals("is_Root", false)
		return c.Next()
	})

	// Swagger API document for Faisu and Vincy
	app.Get("/docs/*", swagger.HandlerDefault)

	// Health
	app.Get("/healthz", func(c *fiber.Ctx) error { return c.SendString("ok") })

	// Routes
	// Auth
	routes.SetupAuth(app, client)

	// Users
	routes.SetupRoutesUser(app, client)

	// Role
	routes.SetupRoutesAbility(app)
	routes.SetupRoutesOrg(app)
	routes.SetupRoutesMembership(app)
	routes.SetupRoutesPosition(app)
	routes.SetupRoutesPolicy(app)

	// Posts
	// routes.SetupRoutesPost(app, client)

	//Events
	routes.SetupRoutesEvent(app)

	// RUN SERVER
	log.Fatal(app.Listen(":" + cfg.Port))
}
