package routes

import (
	"github.com/gofiber/fiber/v2"
	"main-webbase/internal/controllers"
	"main-webbase/internal/repository"
)

func SetupRoutesPolicy(api fiber.Router) {
	repo := repository.NewPolicyRepository()
    h := controllers.NewPolicyHandler(repo)

    policies := api.Group("/policies")
    policies.Post("/", h.CreatePolicy) // POST /api/policies
    policies.Get("/", h.ListPolicies)  // GET /api/policies
}
