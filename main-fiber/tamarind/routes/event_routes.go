package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pllus/main-fiber/tamarind/handlers"
	"github.com/pllus/main-fiber/tamarind/repositories"
	"github.com/pllus/main-fiber/tamarind/services"
)

func RegisterEventRoutes(api fiber.Router) {
	eventRepo := repositories.NewEventRepository()
	mRepo := repositories.NewMembershipRepository()
	eventService := services.NewEventService(eventRepo, mRepo)
	handler := handlers.NewEventHandler(eventService)

	events := api.Group("/event")
	events.Post("/", handler.CreateEvent)
	events.Get("/", handler.GetAllVisible)
	// add Delete later if needed
}
