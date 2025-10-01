package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pllus/main-fiber/tamarind/repositories"
)

type PolicyHandler struct {
	policyRepo *repositories.PolicyRepository
}

func NewPolicyHandler(r *repositories.PolicyRepository) *PolicyHandler {
	return &PolicyHandler{policyRepo: r}
}

func (h *PolicyHandler) CreatePolicy(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "create policy success"})
}

func (h *PolicyHandler) ListPolicies(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"data": []string{"policy1", "policy2"}})
}
