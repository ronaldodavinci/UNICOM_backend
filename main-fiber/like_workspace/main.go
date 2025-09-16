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
	"like_workspace/bootstrap"
	"like_workspace/internal/routes"

	"github.com/gofiber/swagger"
	// "like_workspace/internal/handlers"
)

func main() {
	// --- MongoDB Connection ---
	client := database.ConnectMongo()
	cfg := database.LoadConfig()
	// defer database.DisconnectMongo()

	if err := bootstrap.EnsureLikeIndexes(client.Database("lll_workspace")); err != nil {
		log.Fatalf("ensure indexes failed: %v", err)
	}

	// --- Fiber App Setup ---
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", "66c6248b98c56c39f018e7d1") // 👈 ใช้ ObjectID จริงของ user
		c.Locals("is_Root", false)
		return c.Next()
	})

	app.Use(func(c *fiber.Ctx) error {
		if uid := c.Get("X-User-ID"); uid != "" {
			c.Locals("user_id", uid) // ใส่เป็น hex ของ ObjectID
		}
		if adm := c.Get("X-Is-Admin"); adm == "true" {
			c.Locals("is_admin", true)
		}
		return c.Next()
	})

	// Swagger docs
	app.Get("/docs/*", swagger.HandlerDefault)

	app.Use(func(c *fiber.Ctx) error {
		if uid := c.Get("X-User-ID"); uid != "" {
			c.Locals("user_id", uid)
		}
		if c.Get("X-Is-Root") == "true" {
			c.Locals("is_root", true)
		}
		return c.Next()
	})

	// app.Get("/limit", handlers.GetPostsLimit(client))

	// routes.GetUsersHandler(app, client)

	// app.Post("/postblog", handlers.CreatePostHandler(client))

	// // Register routes
	// routes.RegisterRoutes(app, client)

	// app.Get("/posts/limit-role", handlers.GetPostsLimitrole(client))
	routes.Register(app, routes.Deps{
		Client: client,
	})

	routes.PostRoutes(app, client)

	routes.LikeRoutes(app, client)

	routes.CommentRoutes(app, client)

	log.Printf("listening at http://localhost:%s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
