package controllers

import (
	"github.com/gofiber/fiber/v2"
	"main-webbase/internal/repository"
    "main-webbase/internal/services"
    "main-webbase/dto"
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
func CreatePosition() fiber.Handler {
    return func(c *fiber.Ctx) error {
        var body dto.PositionCreateDTO
        if err := c.BodyParser(&body); err != nil {
            return fiber.NewError(fiber.StatusBadRequest, "invalid body")
        }

        position, policy, err := services.CreatePositionWithPolicy(body, c.Context())
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

        return c.Status(fiber.StatusCreated).JSON(fiber.Map{
            "message":   "create position success",
            "position": fiber.Map{
                "id":        position.ID.Hex(),
                "key":       position.Key,
                "org_path":  position.Scope.OrgPath,
                "status":    position.Status,
                "rank":      position.Rank,
                "createdAt": position.CreatedAt,
            },
            "policy": fiber.Map{
                "id":           policy.ID.Hex(),
                "position_key": policy.PositionKey,
                "scope":        policy.Scope,
                "org_prefix":   policy.OrgPrefix,
                "actions":      policy.Actions,
                "enabled":      policy.Enabled,
                "createdAt":    policy.CreatedAt,
            },
        })
    }
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
    positions, err := h.positionRepo.FindAll(c.Context())
    if err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, err.Error())
    }
    return c.JSON(fiber.Map{"data": positions})
}
