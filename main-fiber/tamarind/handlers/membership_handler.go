package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pllus/main-fiber/tamarind/models"
	"github.com/pllus/main-fiber/tamarind/repositories"
)

type MembershipHandler struct {
	repo *repositories.MembershipRepository
}

func NewMembershipHandler(r *repositories.MembershipRepository) *MembershipHandler {
	return &MembershipHandler{repo: r}
}

func (h *MembershipHandler) CreateMembership(c *fiber.Ctx) error {
	var req models.Membership
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	if err := h.repo.Insert(c.Context(), req); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"message": "membership created", "data": req})
}

func (h *MembershipHandler) ListMemberships(c *fiber.Ctx) error {
	mems, err := h.repo.FindAll(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(mems)
}