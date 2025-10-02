package controllers

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"main-webbase/dto"
	"main-webbase/internal/services"
)

type AbilitiesHandler struct {
	authz *services.AuthzService
}

func NewAbilitiesHandler(a *services.AuthzService) *AbilitiesHandler {
	return &AbilitiesHandler{authz: a}
}

// GetAbilities godoc
// @Summary      Get user abilities
// @Description  Returns allowed actions for the user in the given org_path
// @Tags         abilities
// @Accept       json
// @Produce      json
// @Param        org_path query string true "Organization Path"
// @Success      200 {object} dto.AbilitiesResponse
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /abilities [get]
func (h *AbilitiesHandler) GetAbilities(c *fiber.Ctx) error {
	orgPath := strings.TrimSpace(c.Query("org_path"))
	if orgPath == "" {
		return fiber.NewError(fiber.StatusBadRequest, "org_path required")
	}

	userID, err := services.UserIDFrom(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

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
