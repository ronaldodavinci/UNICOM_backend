package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pllus/main-fiber/tamarind/repositories"
)

type MembershipHandler struct {
	membershipRepo *repositories.MembershipRepository
}

func NewMembershipHandler(r *repositories.MembershipRepository) *MembershipHandler {
	return &MembershipHandler{membershipRepo: r}
}

func (h *MembershipHandler) CreateMembership(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "create membership success"})
}

func (h *MembershipHandler) ListMemberships(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"data": []string{"membership1", "membership2"}})
}
