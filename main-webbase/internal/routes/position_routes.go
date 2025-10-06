package routes

import (
	"github.com/gofiber/fiber/v2"
	"main-webbase/internal/controllers"
	"main-webbase/internal/repository"
)

func SetupRoutesPosition(app *fiber.App) {
	repo := repository.NewPositionRepository()
	h := controllers.NewPositionHandler(repo)

	positions := app.Group("/positions")
	positions.Post("/", controllers.CreatePosition())
	positions.Get("/", h.ListPositions)
}
