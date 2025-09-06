package main

import (
	"log"

	"github.com/gofiber/fiber/v2"

	// "github.com/gofiber/fiber/v2/middleware/cors"
	// "github.com/gofiber/fiber/v2/middleware/logger"
	// "github.com/gofiber/fiber/v2/middleware/recover"
	"like_workspace/database"
	_ "like_workspace/docs"
	// "like_workspace/internal/handlers"
	"like_workspace/internal/routes"

	"github.com/gofiber/swagger"
	// "like_workspace/internal/handlers"
)

func main() {
	// --- MongoDB Connection ---
	client := database.ConnectMongo()
	cfg := database.LoadConfig()
	// defer database.DisconnectMongo()

	// --- Fiber App Setup ---
	app := fiber.New()

	// Swagger docs
	app.Get("/docs/*", swagger.HandlerDefault)


	// app.Get("/limit", handlers.GetPostsLimit(client))

	// routes.GetUsersHandler(app, client)

	// app.Post("/postblog", handlers.CreatePostHandler(client))

	// // Register routes
	// routes.RegisterRoutes(app, client)

	// app.Get("/posts/limit-role", handlers.GetPostsLimitrole(client))
	routes.Register(app, routes.Deps{
		Client: client,
	})

	log.Printf("listening at http://localhost:%s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}