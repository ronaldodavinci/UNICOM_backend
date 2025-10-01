package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pllus/main-fiber/tamarind/handlers"
	"github.com/pllus/main-fiber/tamarind/repositories"
)

func RegisterPostRoutes(api fiber.Router) {
	h := handlers.NewPostHandler(repositories.NewPostRepository())
	posts := api.Group("/posts")
	posts.Post("/", h.CreatePost)
	posts.Get("/", h.ListPosts)
}
