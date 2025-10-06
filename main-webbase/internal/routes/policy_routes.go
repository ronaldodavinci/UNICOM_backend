package routes

import (
	"github.com/gofiber/fiber/v2"
	"main-webbase/internal/controllers"
)

func SetupRoutesPolicy(app fiber.Router) {
	policy := app.Group("/policies")

	policy.Put("/", controllers.UpdatePolicyHandler())
}
