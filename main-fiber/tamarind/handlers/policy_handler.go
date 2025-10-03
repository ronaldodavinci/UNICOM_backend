package handlers

import (
    "github.com/gofiber/fiber/v2"
    "github.com/pllus/main-fiber/tamarind/models"
    "github.com/pllus/main-fiber/tamarind/repositories"
    "time"
)

type PolicyHandler struct {
    repo *repositories.PolicyRepository
}

func NewPolicyHandler(r *repositories.PolicyRepository) *PolicyHandler {
    return &PolicyHandler{repo: r}
}

// POST /api/policies
func (h *PolicyHandler) CreatePolicy(c *fiber.Ctx) error {
    var req models.Policy
    if err := c.BodyParser(&req); err != nil {
        return fiber.NewError(fiber.StatusBadRequest, "invalid body")
    }
    req.CreatedAt = time.Now()

    if err := h.repo.Insert(c.Context(), req); err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, err.Error())
    }

    return c.JSON(fiber.Map{"message": "policy created", "data": req})
}

// GET /api/policies
func (h *PolicyHandler) ListPolicies(c *fiber.Ctx) error {
    policies, err := h.repo.FindAll(c.Context())
    if err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, err.Error())
    }
    return c.JSON(policies)
}