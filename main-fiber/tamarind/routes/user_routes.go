package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pllus/main-fiber/tamarind/handlers"
	"github.com/pllus/main-fiber/tamarind/repositories"
)

func RegisterUserRoutes(api fiber.Router) {
	h := handlers.NewUserHandler(repositories.NewUserRepository())
	users := api.Group("/users")
	users.Post("/", h.CreateUser)
	users.Get("/", h.ListUsers)
}
