package controllers

import (
	"github.com/gofiber/fiber/v2"
    "main-webbase/internal/services"
    "main-webbase/internal/middleware"
    "main-webbase/dto"
    repo "main-webbase/internal/repository"
)

// UpdatePolicyHandler godoc
// @Summary      Update Policy actions
// @Description  Updates policy actions for a position. Only the actions sent will be kept. Supported actions:
//               "membership:assign", "organize:create", "event:create". Sending fewer actions will remove the rest.
// @Tags         Policies
// @Accept       json
// @Produce      json
// @Param        body  body      dto.PolicyUpdateDTO  true  "Policy update data"
// @Success      201   {object}  map[string]interface{} "policy updated successfully"
// @Failure      400   {object}  dto.ErrorResponse "invalid body"
// @Failure      401   {object}  dto.ErrorResponse "unauthorized"
// @Failure      403   {object}  dto.ErrorResponse "no permission to manage this policy"
// @Failure      404   {object}  dto.ErrorResponse "target policy not found"
// @Failure      500   {object}  dto.ErrorResponse "failed to update policy"
// @Router       /policies [put]
func UpdatePolicyHandler() fiber.Handler {
    return func(c *fiber.Ctx) error {
        var body dto.PolicyUpdateDTO
        if err := c.BodyParser(&body); err != nil {
            return fiber.NewError(fiber.StatusBadRequest, "invalid body")
        }

        uid, err := middleware.UIDFromLocals(c)
        if err != nil {
            return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
        }

        userPolicy, err := services.MyUserPolicy(c.Context(), uid)
        if err != nil {
            return fiber.NewError(fiber.StatusNotFound, "target policy not found")
        }

        targetPolicy, err := repo.FindPolicyByKeyandPath(c.Context(), body.Key, body.OrgPath)
        if err != nil {
            return fiber.NewError(fiber.StatusNotFound, "target policy not found")
        }
        if targetPolicy == nil {
			return fiber.NewError(fiber.StatusNotFound, "target policy not found")
		}

        if err := services.CanManagePolicy(userPolicy, targetPolicy); err != nil {
			return fiber.NewError(fiber.StatusForbidden, "no permission to manage this policy")
		}

        targetPolicy.Actions = body.Actions

        if err := services.UpdatedPolicy(c.Context(), targetPolicy); err != nil {
            return fiber.NewError(fiber.StatusInternalServerError, "failed to update policy")
        }
        return c.Status(fiber.StatusCreated).JSON(fiber.Map{
            "message":   "update policy success",
            "policy": fiber.Map{
                "position_key": targetPolicy.PositionKey,
                "org_prefix":   targetPolicy.OrgPrefix,
                "actions":      targetPolicy.Actions,
                "enabled":      targetPolicy.Enabled,
                "createdAt":    targetPolicy.CreatedAt,
            },
        })
    }
}