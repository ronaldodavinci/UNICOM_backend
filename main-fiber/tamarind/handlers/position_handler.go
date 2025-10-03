package handlers

import (
    "github.com/gofiber/fiber/v2"
    "github.com/pllus/main-fiber/tamarind/models"
    "github.com/pllus/main-fiber/tamarind/repositories"
)

type PositionHandler struct {
    positionRepo *repositories.PositionRepository
}

func NewPositionHandler(r *repositories.PositionRepository) *PositionHandler {
    return &PositionHandler{positionRepo: r}
}

func (h *PositionHandler) CreatePosition(c *fiber.Ctx) error {
    var req models.Position
    if err := c.BodyParser(&req); err != nil {
        return fiber.NewError(fiber.StatusBadRequest, "invalid body")
    }

    if err := h.positionRepo.Insert(c.Context(), req); err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, err.Error())
    }

    return c.JSON(fiber.Map{"message": "create position success"})
}

func (h *PositionHandler) ListPositions(c *fiber.Ctx) error {
    positions, err := h.positionRepo.FindAll(c.Context())
    if err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, err.Error())
    }
    return c.JSON(fiber.Map{"data": positions})
}