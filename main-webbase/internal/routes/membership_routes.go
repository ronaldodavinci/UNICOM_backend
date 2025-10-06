package routes

import (
	"github.com/gofiber/fiber/v2"
	"main-webbase/internal/controllers"
	"main-webbase/internal/repository"
)

func SetupRoutesMembership(api fiber.Router) {
	repo := repository.NewMembershipRepository()
	h := controllers.NewMembershipHandler(repo)

	memberships := api.Group("/memberships")
	memberships.Post("/", controllers.CreateMembership())
	memberships.Get("/", h.ListMemberships)
}
