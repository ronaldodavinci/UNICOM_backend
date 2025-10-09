// @title Fiber MongoDB API
// @version 1.0
// @description This is a sample server for user management.
// @host localhost:8000
// @BasePath /

package main

import (
	"log"
	"os"

	_ "main-webbase/docs"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	"github.com/joho/godotenv"

	"main-webbase/config"
	"main-webbase/database"
	"main-webbase/internal/middleware"
	"main-webbase/internal/routes"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ Warning: .env file not found, using system environment variables")
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		panic("JWT_SECRET is required")
	}
	log.Printf("env: JWT_SECRET len=%d", len(os.Getenv("JWT_SECRET")))

	// Load configuration
	cfg := config.LoadConfig()

	// Connect to the database
	client := database.ConnectMongo(cfg.MongoURI, cfg.MongoDB)
	defer client.Disconnect(nil)

	db := client.Database("unicom")

	// Fiber app
	app := fiber.New()

	// app.Use(func(c *fiber.Ctx) error {
	// 	c.Locals("user_id", "68bd6ff6f80438824239b8a9")
	// 	c.Locals("is_Root", false)
	// 	return c.Next()
	// })

	// Swagger API document for Faisu and Vincy
	app.Get("/docs/*", swagger.HandlerDefault)

	// Health
	app.Get("/healthz", func(c *fiber.Ctx) error { return c.SendString("ok") })

	// Get JWT with login
	routes.SetupAuth(app)

	app.Use(middleware.JWTUidOnly(secret))
	app.Use(middleware.InjectViewer(db))

	// Routes
	routes.SetupRoutesUser(app)
	// routes.SetupRoutesAbility(app)
	routes.SetupRoutesOrg(app)
	routes.SetupRoutesMembership(app)
	routes.SetupRoutesPosition(app)
	routes.SetupRoutesPolicy(app)
	// routes.SetupRoutesPost(app, client)
	routes.SetupRoutesEvent(app, client)
	routes.SetupRoutesPost(app, client)
	routes.CommentRoutes(app, client)

	// RUN SERVER
	log.Fatal(app.Listen(":" + cfg.Port))
}
