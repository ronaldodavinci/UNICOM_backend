package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pllus/main-fiber/tamarind/repositories"
)

type PostHandler struct {
	postRepo *repositories.PostRepository
}

func NewPostHandler(r *repositories.PostRepository) *PostHandler {
	return &PostHandler{postRepo: r}
}

func (h *PostHandler) CreatePost(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "create post success"})
}

func (h *PostHandler) ListPosts(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"data": []string{"post1", "post2"}})
}
