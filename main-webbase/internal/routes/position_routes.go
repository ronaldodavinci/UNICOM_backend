package routes

import (
	"github.com/gofiber/fiber/v2"
	"main-webbase/internal/controllers"
)

func SetupRoutesPosition(app *fiber.App) {

	positions := app.Group("/positions")
	positions.Post("/", controllers.CreatePosition())
}
