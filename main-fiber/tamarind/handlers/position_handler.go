package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pllus/main-fiber/tamarind/repositories"
)

type PositionHandler struct {
	positionRepo *repositories.PositionRepository
}

func NewPositionHandler(r *repositories.PositionRepository) *PositionHandler {
	return &PositionHandler{positionRepo: r}
}

func (h *PositionHandler) CreatePosition(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "create position success"})
}

func (h *PositionHandler) ListPositions(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"data": []string{"position1", "position2"}})
}
