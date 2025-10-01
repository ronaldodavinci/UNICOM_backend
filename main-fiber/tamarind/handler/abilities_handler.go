package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pllus/main-fiber/tamarind/dto"
	"github.com/pllus/main-fiber/tamarind/services"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AbilitiesHandler struct {
	authz *services.AuthzService
}

func NewAbilitiesHandler(a *services.AuthzService) *AbilitiesHandler {
	return &AbilitiesHandler{authz: a}
}

// GET /api/abilities
func (h *AbilitiesHandler) GetAbilities(c *fiber.Ctx) error {
	orgPath := strings.TrimSpace(c.Query("org_path"))
	if orgPath == "" {
		return fiber.NewError(fiber.StatusBadRequest, "org_path required")
	}

	userIDHex := c.Locals("user_id").(string)
	userID, _ := primitive.ObjectIDFromHex(userIDHex)

	actions := []string{
		"membership:assign", "membership:revoke",
		"position:create", "policy:write",
		"event:create", "event:manage",
		"post:create", "post:moderate",
	}

	allowed, err := h.authz.AbilitiesFor(c.Context(), userID, orgPath, actions)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "abilities failed")
	}
	return c.JSON(dto.AbilitiesResponse{OrgPath: orgPath, Abilities: allowed, Version: "pol-v2"})
}
