package routes

import (
	"main-webbase/internal/controllers"

	"github.com/gofiber/fiber/v2"
)

func SetupAuth(app *fiber.App) {
	app.Post("/register", func(c *fiber.Ctx) error {
		return controllers.Register(c)
	})

	app.Post("/login", func(c *fiber.Ctx) error {
		return controllers.Login(c)
	})
}
