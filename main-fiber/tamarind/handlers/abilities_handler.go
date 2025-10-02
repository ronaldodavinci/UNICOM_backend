package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pllus/main-fiber/tamarind/services"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AbilitiesHandler struct {
	authzService *services.AuthzService
}

func NewAbilitiesHandler(s *services.AuthzService) *AbilitiesHandler {
	return &AbilitiesHandler{authzService: s}
}

// GET /api/abilities/:userId
func (h *AbilitiesHandler) GetAbilities(c *fiber.Ctx) error {
	userIDHex := c.Params("userId")
	userID, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid user id")
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