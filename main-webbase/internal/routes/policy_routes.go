package routes

import (
	"github.com/gofiber/fiber/v2"
	"main-webbase/internal/controllers"
	"main-webbase/internal/repository"
)

func SetupRoutesPolicy(api fiber.Router) {
	h := controllers.NewPolicyHandler(repository.NewPolicyRepository())
	policies := api.Group("/policies")
	policies.Post("/", h.CreatePolicy)
	policies.Get("/", h.ListPolicies)
}
