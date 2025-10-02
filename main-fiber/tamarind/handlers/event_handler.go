package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pllus/main-fiber/tamarind/models"
	"github.com/pllus/main-fiber/tamarind/repositories"
	"go.mongodb.org/mongo-driver/bson"
)

type EventHandler struct {
	repo *repositories.EventRepository
}

func NewEventHandler(r *repositories.EventRepository) *EventHandler {
	return &EventHandler{repo: r}
}

// POST /api/events
func (h *EventHandler) CreateEvent(c *fiber.Ctx) error {
	var req models.Event
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}

	// insert event
	if err := h.repo.InsertEvent(c.Context(), req); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{"message": "event created", "data": req})
}

// GET /api/events
func (h *EventHandler) ListEvents(c *fiber.Ctx) error {
	events, err := h.repo.FindEvents(c.Context(), bson.M{})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(events)
}