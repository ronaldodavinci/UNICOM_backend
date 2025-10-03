package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pllus/main-fiber/tamarind/handlers"
	"github.com/pllus/main-fiber/tamarind/repositories"
)

func RegisterUserRoutes(api fiber.Router) {
	repo := repositories.NewUserRepository()
	h := handlers.NewUserHandler(repo)

	users := api.Group("/users")
	users.Post("/", h.CreateUser)  // POST /api/users
	users.Get("/", h.ListUsers)    // GET /api/users
	users.Get("/:id", h.GetUser)   // GET /api/users/:id
	users.Put("/:id", h.UpdateUser) // PUT /api/users/:id
	users.Delete("/:id", h.DeleteUser) // DELETE /api/users/:id
}