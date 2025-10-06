package routes

import (
	"github.com/gofiber/fiber/v2"
	"main-webbase/internal/controllers"
)

func SetupRoutesOrg(app *fiber.App) {

	org := app.Group("/org/units")

	org.Post("/", controllers.CreateOrgUnitHandler())
	org.Get("/tree", controllers.GetOrgTree())
}
