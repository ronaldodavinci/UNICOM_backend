package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	_ "github.com/pllus/main-fiber/docs" // ถ้ามี swagger docs

	"github.com/pllus/main-fiber/lukkate_workspace/internal/db"
	"github.com/pllus/main-fiber/lukkate_workspace/internal/handlers"
)

func main() {
	app := fiber.New()

	// Swagger ตามเดิม
	app.Get("/docs/*", swagger.HandlerDefault)

	// --- Connect Mongo ---
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		// ควรเปลี่ยนในงานจริง
		mongoURI = "mongodb+srv://root:971397@cluster01.wawl1f9.mongodb.net/"
	}
	if err := db.Connect(mongoURI); err != nil {
		log.Fatal(err)
	}
	defer db.Disconnect()

	// --- Routes ---
	app.Get("/hello", func(c *fiber.Ctx) error { return c.SendString("Hello, World!") })
	app.Get("/Post", handlers.GetUsers)

	// --- Server ---
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Fatal(app.Listen(":" + port))
}
