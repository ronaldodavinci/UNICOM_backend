package controllers

import (
	"github.com/gofiber/fiber/v2"
	// "main-webbase/internal/repository"
    "main-webbase/internal/services"
    "main-webbase/dto"
)

// CreatePosition godoc
// @Summary      Create a new Position with Policy
// @Description  Create a new position and attach a policy. Policy actions support only:
//               "membership:assign", "organize:create", "event:create".
//               Any missing action in update will be removed from the policy.
// @Tags         Positions
// @Accept       json
// @Produce      json
// @Param        body  body      dto.PositionCreateDTO  true  "Position & Policy data"
// @Success      201   {object}  map[string]interface{} "position created successfully"
// @Failure      400   {object}  dto.ErrorResponse "invalid body"
// @Failure      500   {object}  dto.ErrorResponse "internal server error"
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

// ListPositions 
// @Summary      List positions
// @Description  Returns a list of positions
// @Tags         positions
// @Accept       json
// @Produce      json
// @Success      200 {object} map[string][]string
// @Failure      500 {object} map[string]interface{}
// @Router       /positions [get]
// func (h *PositionHandler) ListPositions(c *fiber.Ctx) error {
//     positions, err := h.positionRepo.FindAll(c.Context())
//     if err != nil {
//         return fiber.NewError(fiber.StatusInternalServerError, err.Error())
//     }
//     return c.JSON(fiber.Map{"data": positions})
// }
