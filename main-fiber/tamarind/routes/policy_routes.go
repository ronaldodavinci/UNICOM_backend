package routes

import (
    "github.com/gofiber/fiber/v2"
    "github.com/pllus/main-fiber/tamarind/handlers"
    "github.com/pllus/main-fiber/tamarind/repositories"
)

func RegisterPolicyRoutes(api fiber.Router) {
    repo := repositories.NewPolicyRepository()
    h := handlers.NewPolicyHandler(repo)

    policies := api.Group("/policies")
    policies.Post("/", h.CreatePolicy) // POST /api/policies
    policies.Get("/", h.ListPolicies)  // GET /api/policies
}