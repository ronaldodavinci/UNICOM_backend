package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pllus/main-fiber/tamarind/handlers"
	"github.com/pllus/main-fiber/tamarind/repositories"
)

func RegisterEventRoutes(api fiber.Router) {
	repo := repositories.NewEventRepository()
	h := handlers.NewEventHandler(repo)

	events := api.Group("/events")
	events.Post("/", h.CreateEvent)
	events.Get("/", h.ListEvents)
}