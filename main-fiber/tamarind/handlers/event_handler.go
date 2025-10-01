package handlers

import (

	"github.com/gofiber/fiber/v2"
	"github.com/pllus/main-fiber/tamarind/dto"
	"github.com/pllus/main-fiber/tamarind/models"
	"github.com/pllus/main-fiber/tamarind/services"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EventHandler struct {
	eventService *services.EventService
}

func NewEventHandler(e *services.EventService) *EventHandler {
	return &EventHandler{eventService: e}
}

// POST /api/event
func (h *EventHandler) CreateEvent(c *fiber.Ctx) error {
	var req dto.EventRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}

	nodeID, err := primitive.ObjectIDFromHex(req.NodeID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid node_id")
	}

	event := models.Event{
		NodeID:           nodeID,
		Topic:            req.Topic,
		Description:      req.Description,
		MaxParticipation: req.MaxParticipation,
		OrgOfContent:     req.OrgOfContent,
		Status:           req.Status,
	}
	var schedules []models.EventSchedule
	for _, s := range req.Schedules {
		schedules = append(schedules, models.EventSchedule{
			ID:          primitive.NewObjectID(),
			EventID:     event.ID,
			Date:        s.Date,
			TimeStart:   s.TimeStart,
			TimeEnd:     s.TimeEnd,
			Location:    &s.Location,
			Description: &s.Description,
		})
	}

	if err := h.eventService.CreateEventWithSchedules(c.Context(), event, schedules); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	var reports []dto.EventScheduleReport
	for _, sc := range schedules {
		reports = append(reports, dto.EventScheduleReport{
			Date: sc.Date, StartTime: sc.TimeStart, EndTime: sc.TimeEnd,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(dto.EventReport{
		EventID:    event.ID.Hex(),
		EventTopic: event.Topic,
		Schedules:  reports,
	})
}

// GET /api/event
func (h *EventHandler) GetAllVisible(c *fiber.Ctx) error {
	userIDHex := c.Locals("user_id").(string)
	viewerID, _ := primitive.ObjectIDFromHex(userIDHex)

	// TODO: should call AudienceService/AllUserOrg here
	orgs := []string{} // placeholder

	events, err := h.eventService.GetVisibleEvents(c.Context(), viewerID, orgs)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(events)
}
