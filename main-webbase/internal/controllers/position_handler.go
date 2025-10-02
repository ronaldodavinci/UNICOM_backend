package controllers

import (
	"github.com/gofiber/fiber/v2"
	"main-webbase/internal/repository"
)

type PositionHandler struct {
	positionRepo *repository.PositionRepository
}

func NewPositionHandler(r *repository.PositionRepository) *PositionHandler {
	return &PositionHandler{positionRepo: r}
}

// CreatePosition godoc
// @Summary      Create a new position
// @Description  Adds a new position to the system
// @Tags         positions
// @Accept       json
// @Produce      json
// @Param        position body models.Position true "Position data"
// @Success      200 {object} map[string]string
// @Failure      400 {object} map[string]interface{}
// @Failure      500 {object} map[string]interface{}
// @Router       /positions [post]
func (h *PositionHandler) CreatePosition(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "create position success"})
}

// ListPositions godoc
// @Summary      List positions
// @Description  Returns a list of positions
// @Tags         positions
// @Accept       json
// @Produce      json
// @Success      200 {object} map[string][]string
// @Failure      500 {object} map[string]interface{}
// @Router       /positions [get]
func (h *PositionHandler) ListPositions(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"data": []string{"position1", "position2"}})
}
