package routes

import (
	"github.com/gofiber/fiber/v2"
	"main-webbase/internal/controllers"
)

func SetupRoutesMembership(api fiber.Router) {

    memberships := api.Group("/memberships")
    memberships.Post("/", controllers.CreateMembership())
    memberships.Get("/users", controllers.ListMembershipsWithUsers())
    memberships.Patch("/:id", controllers.DeactivateMembership())
}
