package main

import (
	"log"
	"os"

	
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	_ "like_workspace/docs"

	"like_workspace/database"
	"like_workspace/internal/routes"
	"like_workspace/internal/handlers"
)

func main() {
	// --- MongoDB Connection ---
	client := configs.ConnectMongo()
	// defer database.DisconnectMongo()

	// --- Fiber App Setup ---
	app := fiber.New()

	// Swagger docs
	app.Get("/docs/*", swagger.HandlerDefault)

	app.Get("/db", handlers.GetPostsLimit(client))


	app.Get("/Post", routes.GetUsersHandler(client))

	app.Post("/postblog", handlers.CreatePostHandler(client))

	// Register routes
	routes.RegisterRoutes(app, client)

	// --- Server variables ---
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Fatal(app.Listen(":" + port))
}
