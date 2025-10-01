package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pllus/main-fiber/tamarind/handlers"
	"github.com/pllus/main-fiber/tamarind/repositories"
)

func RegisterPolicyRoutes(api fiber.Router) {
	h := handlers.NewPolicyHandler(repositories.NewPolicyRepository())
	policies := api.Group("/policies")
	policies.Post("/", h.CreatePolicy)
	policies.Get("/", h.ListPolicies)
}
