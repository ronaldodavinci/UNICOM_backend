package controllers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"

	"main-webbase/database"
	"main-webbase/dto"
	"main-webbase/internal/middleware"
	"main-webbase/internal/services"
)

// CreateEventHandler godoc
// @Summary Create new event
// @Description Create an event with schedules
// @Tags events
// @Accept json
// @Produce json
// @Param event body dto.EventRequestDTO true "Event payload"
// @Success 201 {object} dto.EventReport
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /event [post]
func CreateEventHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body dto.EventRequestDTO
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		if body.NodeID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "NodeID is required"})
		}

		if !canPostAs(viewerFrom(c), body.PostedAs.OrgPath, body.PostedAs.PositionKey) {
			return c.Status(fiber.StatusForbidden).
				JSON(dto.ErrorResponse{Error: "forbidden: you cannot post as this role1"})
		}

		event, schedules, err := services.CreateEventWithSchedules(body, c.Context())
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		reportSchedules := make([]dto.EventScheduleReport, len(schedules))
		for i, s := range schedules {
			reportSchedules[i] = dto.EventScheduleReport{
				Date:      s.Date,
				StartTime: s.Time_start,
				EndTime:   s.Time_end,
			}
		}

		return c.Status(fiber.StatusCreated).JSON(dto.EventReport{
			EventID:    event.ID.Hex(),
			EventTopic: event.Topic,
			Schedules:  reportSchedules,
		})
	}
}

// GetAllVisibleEventHandler godoc
// @Summary Get all visible events
// @Description Return all events that the current user can see
// @Tags events
// @Produce json
// @Success 200 {array} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /event [get]
func GetAllVisibleEventHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		viewerID, err := services.UserIDFrom(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
		}
		orgSets, err := services.AllUserOrg(viewerID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get user orgs"})
		}

		events, err := services.GetVisibleEvents(viewerID, ctx, orgSets)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(events)
	}
}

func GetEventDetailHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		eventID := c.Params("value")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		event, err := services.GetEventDetail(eventID, ctx)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(event)
	}
}

func ParticipateEventWithNoFormHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		eventID := c.Params("value")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		uid, err := middleware.UIDFromLocals(c)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
		}

		if err = services.ParticipateEvent(eventID, uid, ctx); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message":  "Participation successful",
			"event_id": eventID,
			"user_id":  uid,
		})
	}
}

// DeleteEventHandler godoc
// @Summary Soft delete event
// @Description Mark event as hidden
// @Tags events
// @Param id path string true "Event ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /event/{id} [delete]
func DeleteEventHandler(c *fiber.Ctx) error {
	collection_event := database.DB.Collection("event")

	idHex := c.Params("id")
	event_id, err := bson.ObjectIDFromHex(idHex)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = collection_event.UpdateOne(ctx, bson.M{"_id": event_id},
		bson.M{"$set": bson.M{"status": "hidden", "updated_at": time.Now()}})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete event"})
	}

	return c.JSON(fiber.Map{"message": "Event deleted (soft)"})
}

// ยังไม่เสร็จ
// func UpdateEventHandler(c *fiber.Ctx) error {
// 	collection_event := client.Database("big_workspace").Collection("event")

// 	idHex := c.Params("id")
// 	eventID, err := bson.ObjectIDFromHex(idHex)
//     if err != nil {
//         return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
//     }

//     var req map[string]interface{}
//     if err := c.BodyParser(&req); err != nil {
//         return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
//     }

//     req["updated_at"] = time.Now()

//     ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//     defer cancel()

//     _, err = collection_event.UpdateOne(ctx,
//         bson.M{"_id": eventID},
//         bson.M{"$set": req})
//     if err != nil {
//         return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update event"})
//     }

//     return c.JSON(fiber.Map{"message": "Event updated"})
// }
