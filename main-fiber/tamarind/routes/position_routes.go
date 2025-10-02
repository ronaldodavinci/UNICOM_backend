package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pllus/main-fiber/tamarind/handlers"
	"github.com/pllus/main-fiber/tamarind/repositories"
)

func RegisterPositionRoutes(api fiber.Router) {
	repo := repositories.NewPositionRepository()
	h := handlers.NewPositionHandler(repo)

	positions := api.Group("/positions")
	positions.Post("/", h.CreatePosition)
	positions.Get("/", h.ListPositions)
}
