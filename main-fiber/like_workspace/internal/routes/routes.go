package routes

import (
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"like_workspace/internal/handlers"
)

func RegisterRoutes(app *fiber.App, client *mongo.Client) {
	app.Get("/hello", handlers.GetHello)
	app.Get("/Post", func(c *fiber.Ctx) error {
		return handlers.GetUserHandler(c, client)
	})
}
