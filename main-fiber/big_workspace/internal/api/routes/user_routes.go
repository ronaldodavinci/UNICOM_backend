package routes

import (
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"main-fiber/big_workspace/internal/api/controllers"
)

func SetupRoutesUser(app *fiber.App, client *mongo.Client) {
	app.Get("/hello", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	app.Post("/users", func(c *fiber.Ctx) error {
		return controllers.CreateUser(c, client)
	})

	app.Get("/users", func(c *fiber.Ctx) error {
		return controllers.GetAllUser(c, client)
	})

	app.Get("/users/:id", func(c *fiber.Ctx) error {
		return controllers.GetUserByID(c, client)
	})

	app.Delete("/users/:id", func(c *fiber.Ctx) error {
		return controllers.DeleteUser(c, client)
	})
}
