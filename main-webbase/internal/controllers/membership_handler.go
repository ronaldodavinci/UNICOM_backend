package controllers

import (
	"github.com/gofiber/fiber/v2"
	"main-webbase/internal/models"
	repo "main-webbase/internal/repository"
)

// CreateMembership godoc
// @Summary      Create a new membership
// @Description  Assigns a user to an organization and position
// @Tags         Memberships
// @Accept       json
// @Produce      json
// @Param        body  body      models.Membership  true  "Membership data"
// @Success      200   {object}  models.Membership "membership created"
// @Failure      400   {object}  dto.ErrorResponse "invalid body"
// @Failure      500   {object}  dto.ErrorResponse "internal server error"
// @Router       /memberships [post]
func CreateMembership() fiber.Handler {
	return func (c *fiber.Ctx) error {
		var req models.Membership
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid body")
		}
		if err := repo.InsertMembership(c.Context(), req); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(fiber.Map{"message": "membership created", "data": req})
	}
}

// ListMemberships 
// @Summary      List memberships
// @Description  Returns a list of memberships
// @Tags         memberships
// @Accept       json
// @Produce      json
// @Success      200 {object} map[string][]string
// @Failure      500 {object} map[string]interface{}
// @Router       /memberships [get]
// func (h *MembershipHandler) ListMemberships(c *fiber.Ctx) error {
// 	mems, err := h.membershipRepo.FindAll(c.Context())
// 	if err != nil {
// 		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
// 	}
// 	return c.JSON(mems)
// }
