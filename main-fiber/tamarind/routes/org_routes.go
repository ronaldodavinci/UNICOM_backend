package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pllus/main-fiber/tamarind/handlers"
	"github.com/pllus/main-fiber/tamarind/repositories"
)

func RegisterOrgRoutes(api fiber.Router) {
	orgRepo := repositories.NewOrgUnitRepository()
	handler := handlers.NewOrgTreeHandler(orgRepo)

	org := api.Group("/org/units")
	org.Get("/tree", handler.GetTree)
}
