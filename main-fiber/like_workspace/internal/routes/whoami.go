// internal/routes/whoami_routes.go
package routes

import "github.com/gofiber/fiber/v2"

func WhoAmIRoutes(app *fiber.App) {
	app.Get("/whoami", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"user_id": c.Locals("user_id"),
		})
	})
}
