package routes

import (
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"big_workspace/internal/api/handlers"
)

func SetupRoutes(app *fiber.App, client *mongo.Client) {
	app.Get("/hello", func(c *fiber.Ctx) error {
		return c.sendString("Hello, World!")
	})

	app.Get("/users", func(c *fiber.Ctx) error {
		return handlers.GetUser(c, client)
	})
}