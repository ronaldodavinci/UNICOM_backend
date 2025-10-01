package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pllus/main-fiber/tamarind/handlers"
	"github.com/pllus/main-fiber/tamarind/repositories"
)

func RegisterAuthRoutes(api fiber.Router) {
	userRepo := repositories.NewUserRepository()
	handler := handlers.NewAuthHandler(userRepo, []byte("secret")) // replace with env secret

	auth := api.Group("/auth")
	auth.Post("/login", handler.Login)
	auth.Get("/me", handler.Me)
}
