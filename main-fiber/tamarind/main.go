package main

import (
	"log"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/swagger"

	_ "github.com/pllus/main-fiber/tamarind/docs"
	"github.com/pllus/main-fiber/tamarind/config"
	"github.com/pllus/main-fiber/tamarind/routes"
)

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	app := fiber.New()

	// CORS
	allowed := getEnv("FRONTEND_ORIGINS", "http://localhost:5173")
	app.Use(cors.New(cors.Config{
		AllowOrigins:     strings.Join(strings.Split(allowed, ","), ","),
		AllowMethods:     "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		ExposeHeaders:    "Authorization",
		AllowCredentials: true,
	}))

	// DB
	config.ConnectMongo()

	// Health
	app.Get("/healthz", func(c *fiber.Ctx) error { return c.SendString("ok") })

	// API group
	apiGroup := app.Group("/api")

	// Register all routes
	routes.SetupRoutes(apiGroup)

	// Swagger
	app.Get("/docs/*", swagger.HandlerDefault)

	port := getEnv("PORT", "3000")
	log.Fatal(app.Listen(":" + port))
}
