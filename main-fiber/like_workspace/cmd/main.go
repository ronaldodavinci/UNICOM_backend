package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	_ "like_workspace/docs"

	"like_workspace/configs"
	"like_workspace/internal/routes"
)

func main() {
	// --- MongoDB Connection ---
	client := configs.ConnectMongo()
	defer configs.DisconnectMongo(client)

	// --- Fiber App Setup ---
	app := fiber.New()

	// Swagger docs
	app.Get("/docs/*", swagger.HandlerDefault)

	// 
	app.Get("/Post", routes.GetUsersHandler(client))

	// Register routes
	routes.RegisterRoutes(app, client)

	// --- Server variables ---
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Fatal(app.Listen(":" + port))
}
