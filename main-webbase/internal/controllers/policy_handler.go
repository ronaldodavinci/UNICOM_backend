package controllers

import (
    "time"

	"github.com/gofiber/fiber/v2"
	"main-webbase/internal/repository"
    "main-webbase/internal/models"
)

type PolicyHandler struct {
    repo *repository.PolicyRepository
}

func NewPolicyHandler(r *repository.PolicyRepository) *PolicyHandler {
    return &PolicyHandler{repo: r}
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
    policies, err := h.repo.FindAll(c.Context())
    if err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, err.Error())
    }
    return c.JSON(policies)
}
