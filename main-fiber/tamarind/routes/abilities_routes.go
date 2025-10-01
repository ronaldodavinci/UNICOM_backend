package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pllus/main-fiber/tamarind/handlers"
	"github.com/pllus/main-fiber/tamarind/repositories"
	"github.com/pllus/main-fiber/tamarind/services"
)

func RegisterAbilitiesRoutes(api fiber.Router) {
	mRepo := repositories.NewMembershipRepository()
	pRepo := repositories.NewPolicyRepository()
	authzService := services.NewAuthzService(mRepo, pRepo)
	handler := handlers.NewAbilitiesHandler(authzService)

	abilities := api.Group("/abilities")
	abilities.Get("/", handler.GetAbilities)
}
