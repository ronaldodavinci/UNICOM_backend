package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pllus/main-fiber/tamarind/repositories"
)

type UserHandler struct {
	userRepo *repositories.UserRepository
}

func NewUserHandler(r *repositories.UserRepository) *UserHandler {
	return &UserHandler{userRepo: r}
}

func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "create user success"})
}

func (h *UserHandler) ListUsers(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"data": []string{"user1", "user2"}})
}
