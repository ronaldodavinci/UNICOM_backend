package controllers

import (
	"github.com/gofiber/fiber/v2"
	"main-webbase/internal/repository"
)

type PolicyHandler struct {
	policyRepo *repository.PolicyRepository
}

func NewPolicyHandler(r *repository.PolicyRepository) *PolicyHandler {
	return &PolicyHandler{policyRepo: r}
}

// CreatePolicy godoc
// @Summary      Create a new policy
// @Description  Adds a new policy to the system
// @Tags         policies
// @Accept       json
// @Produce      json
// @Param        policy body models.Role true "Policy data"
// @Success      200 {object} map[string]string
// @Failure      400 {object} map[string]interface{}
// @Failure      500 {object} map[string]interface{}
// @Router       /policies [post]
func (h *PolicyHandler) CreatePolicy(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "create policy success"})
}

// ListPolicies godoc
// @Summary      List policies
// @Description  Returns a list of policies
// @Tags         policies
// @Accept       json
// @Produce      json
// @Success      200 {object} map[string][]string
// @Failure      500 {object} map[string]interface{}
// @Router       /policies [get]
func (h *PolicyHandler) ListPolicies(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"data": []string{"policy1", "policy2"}})
}
