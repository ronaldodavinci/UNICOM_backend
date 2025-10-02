package routes

import (
	"main-webbase/internal/controllers"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func SetupAuth(app *fiber.App, client *mongo.Client) {
	app.Post("/register", func(c *fiber.Ctx) error {
		return controllers.Register(c, client)
	})

	app.Post("/login", func(c *fiber.Ctx) error {
		return controllers.Login(c, client)
	})
}
