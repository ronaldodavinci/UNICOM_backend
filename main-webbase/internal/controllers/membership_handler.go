package controllers

import (
	"github.com/gofiber/fiber/v2"
	"main-webbase/internal/repository"
)

type MembershipHandler struct {
	membershipRepo *repository.MembershipRepository
}

func NewMembershipHandler(r *repository.MembershipRepository) *MembershipHandler {
	return &MembershipHandler{membershipRepo: r}
}

// CreateMembership godoc
// @Summary      Create a new membership
// @Description  Adds a new membership for a user in an organization
// @Tags         memberships
// @Accept       json
// @Produce      json
// @Param        membership body models.Membership true "Membership data"
// @Success      200 {object} map[string]string
// @Failure      400 {object} map[string]interface{}
// @Failure      500 {object} map[string]interface{}
// @Router       /memberships [post]
func (h *MembershipHandler) CreateMembership(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "create membership success"})
}

// ListMemberships godoc
// @Summary      List memberships
// @Description  Returns a list of memberships
// @Tags         memberships
// @Accept       json
// @Produce      json
// @Success      200 {object} map[string][]string
// @Failure      500 {object} map[string]interface{}
// @Router       /memberships [get]
func (h *MembershipHandler) ListMemberships(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"data": []string{"membership1", "membership2"}})
}
