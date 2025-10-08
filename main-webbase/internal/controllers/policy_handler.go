package controllers

import (
    "github.com/gofiber/fiber/v2"
    "main-webbase/internal/services"
    "main-webbase/internal/middleware"
    "main-webbase/dto"
    repo "main-webbase/internal/repository"
    "go.mongodb.org/mongo-driver/v2/bson"
    "strings"
    "main-webbase/database"
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

// ListPolicies godoc
// @Summary      List policies
// @Description  Returns policies, optionally filtered by org_prefix and/or position_key
// @Tags         Policies
// @Accept       json
// @Produce      json
// @Param        org_prefix   query string false "Org prefix filter"
// @Param        position_key query string false "Position key filter"
// @Success      200 {array}  map[string]interface{}
// @Failure      500 {object} map[string]string
// @Router       /policies [get]
func ListPolicies() fiber.Handler {
    return func(c *fiber.Ctx) error {
        r := repo.NewPolicyRepository()
        list, err := r.FindAll(c.Context())
        if err != nil {
            return fiber.NewError(fiber.StatusInternalServerError, err.Error())
        }
        orgPrefix := c.Query("org_prefix")
        posKey := c.Query("position_key")
        out := make([]fiber.Map, 0, len(list))
        for _, p := range list {
            if orgPrefix != "" && !strings.HasPrefix(p.OrgPrefix, orgPrefix) { continue }
            if posKey != "" && p.PositionKey != posKey { continue }
            out = append(out, fiber.Map{
                "_id":          p.ID.Hex(),
                "position_key": p.PositionKey,
                "scope":        p.Scope,
                "org_prefix":   p.OrgPrefix,
                "actions":      p.Actions,
                "enabled":      p.Enabled,
                "created_at":   p.CreatedAt,
                "updatedAt":    p.UpdatedAt,
            })
        }
        return c.JSON(out)
    }
}

// DeletePolicyHandler godoc
// @Summary      Delete policies
// @Description  Delete policies by org_prefix and optional position_key
// @Tags         Policies
// @Accept       json
// @Produce      json
// @Param        org_prefix   query string true  "Org prefix"
// @Param        position_key query string false "Position key"
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /policies [delete]
func DeletePolicyHandler() fiber.Handler {
    return func(c *fiber.Ctx) error {
        orgPrefix := c.Query("org_prefix")
        if orgPrefix == "" {
            return fiber.NewError(fiber.StatusBadRequest, "org_prefix is required")
        }
        posKey := c.Query("position_key")
        col := database.DB.Collection("policies")
        filter := bson.M{"org_prefix": orgPrefix}
        if posKey != "" { filter["position_key"] = posKey }
        res, err := col.DeleteMany(c.Context(), filter)
        if err != nil {
            return fiber.NewError(fiber.StatusInternalServerError, err.Error())
        }
        return c.JSON(fiber.Map{"deleted": res.DeletedCount})
    }
}
