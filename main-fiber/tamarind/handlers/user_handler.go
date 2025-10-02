package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pllus/main-fiber/tamarind/repositories"
	"github.com/pllus/main-fiber/tamarind/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
	"context"
	"go.mongodb.org/mongo-driver/bson"
)

type UserHandler struct {
	repo *repositories.UserRepository
}

func NewUserHandler(r *repositories.UserRepository) *UserHandler {
	return &UserHandler{repo: r}
}

// POST /api/users
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var u models.User
	if err := c.BodyParser(&u); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}
	u.ID = primitive.NewObjectID()
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()

	_, err := h.repo.Insert(context.Background(), u)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(u)
}

// GET /api/users
func (h *UserHandler) ListUsers(c *fiber.Ctx) error {
	users, err := h.repo.FindAll(context.Background())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(users)
}

// GET /api/users/:id
func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	id, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}
	user, err := h.repo.FindByID(context.Background(), id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "not found"})
	}
	return c.JSON(user)
}

// PUT /api/users/:id
func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	id, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}
	var update bson.M
	if err := c.BodyParser(&update); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}
	update["updatedAt"] = time.Now()

	if err := h.repo.Update(context.Background(), id, update); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true})
}

// DELETE /api/users/:id
func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	id, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}
	if err := h.repo.Delete(context.Background(), id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true})
}