package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pllus/main-fiber/tamarind/handlers"
	"github.com/pllus/main-fiber/tamarind/repositories"
	"github.com/pllus/main-fiber/tamarind/services"
)


func RegisterMembershipRoutes(api fiber.Router) {
	repo := repositories.NewMembershipRepository()
	h := handlers.NewMembershipHandler(repo)

	memberships := api.Group("/memberships")
	memberships.Post("/", h.CreateMembership)
	memberships.Get("/", h.ListMemberships)
}
