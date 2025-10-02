package routes

import (
	"github.com/gofiber/fiber/v2"
	"main-webbase/internal/controllers"
	"main-webbase/internal/repository"
)

func SetupRoutesPosition(api fiber.Router) {
	repo := repository.NewPositionRepository()
	h := controllers.NewPositionHandler(repo)

	positions := api.Group("/positions")
	positions.Post("/", h.CreatePosition)
	positions.Get("/", h.ListPositions)
}
