package routes

import (
	"github.com/gofiber/fiber/v2"
)

func RegisterUserRoutes(api fiber.Router) {
	users := api.Group("/users")

	// TODO: wire user handler functions here
	users.Get("/", func(c *fiber.Ctx) error { return c.SendString("list users") })
}
