package controllers

import (
	"github.com/gofiber/fiber/v2"

	"main-webbase/internal/services"
)

type AbilitiesHandler struct {
	authzService *services.AuthzService
}

func NewAbilitiesHandler(s *services.AuthzService) *AbilitiesHandler {
	return &AbilitiesHandler{authzService: s}
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
	userID, err := services.UserIDFrom(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// กำหนด action ที่ต้องการตรวจสอบ เช่น membership, position, event, post ฯลฯ
	actions := []string{
		"membership:assign",
		"membership:revoke",
		"position:create",
		"policy:write",
		"event:create",
		"event:manage",
		"post:create",
		"post:moderate",
	}

	result, err := h.authzService.AbilitiesFor(c.Context(), userID, "", actions)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(result)
}
