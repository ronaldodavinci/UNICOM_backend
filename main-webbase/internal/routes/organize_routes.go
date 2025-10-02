package routes

import (
	"github.com/gofiber/fiber/v2"
	"main-webbase/internal/controllers"
	"main-webbase/internal/repository"
)

func SetupRoutesOrg(app *fiber.App) {
	orgRepo := repository.NewOrgUnitRepository()
	handler := controllers.NewOrgTreeHandler(orgRepo)

	org := app.Group("/org/units")
	org.Get("/tree", handler.GetTree)
}
