package controllers

import (
	"context"
	// "encoding/json"
	"fmt"
	"log"
	"path/filepath"
	// "strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"

	"main-webbase/database"
	"main-webbase/dto"
	"main-webbase/internal/middleware"
	"main-webbase/internal/models"
	"main-webbase/internal/services"
)

// CreateEventHandler godoc
// @Summary Create a new event
// @Description Create an event with optional image upload and schedule creation
// @Tags events
// @Accept multipart/form-data
// @Produce json
//
// @Param file formData file false "Upload event image"
// @Param NodeID formData string true "Node ID"
// @Param topic formData string false "Event topic"
// @Param description formData string false "Event description"
// @Param org_of_content formData string false "Organization responsible for content (e.g. /fac/eng/com)"
// @Param status formData string false "Event status" Enums(active, draft, inactive)
// @Param max_participation formData int false "Maximum number of participants"
// @Param have_form formData bool false "Whether this event has an associated form (true/false)"
// @Param postedAs.org_path formData string true "Organization path of the posting role"
// @Param postedAs.position_key formData string true "Position key of the posting role"
// @Param visibility formData string false "Visibility settings (as JSON string). Example: {\"access\":\"public\"}"
// @Param schedules formData string false "Schedules as JSON array (optional). Example: [{\"date\":\"2025-10-15T00:00:00Z\",\"time_start\":\"2025-10-15T09:00:00Z\",\"time_end\":\"2025-10-15T12:00:00Z\",\"location\":\"Innovation Building Room 301\",\"description\":\"Morning session\"}]"
//
// @Success 201 {object} dto.EventCreateResult "Event created successfully"
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 403 {object} dto.ErrorResponse "Forbidden: You cannot post as this role"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /event [post]
func CreateEventHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body dto.EventRequestDTO

		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).
				JSON(fiber.Map{"error": "invalid request body"})
		}

		// --- optional file upload ---
		file, err := c.FormFile("file")
		if err == nil && file != nil {
			timestamp := time.Now().UnixNano() / 1e6
			ext := filepath.Ext(file.Filename)
			filename := fmt.Sprintf("event_%d%s", timestamp, ext)
			savePath := filepath.Join("/var/www/html/uploads", filename)

			if err := c.SaveFile(file, savePath); err != nil {
				return c.Status(fiber.StatusInternalServerError).
					JSON(fiber.Map{"error": "failed to save file"})
			}

			publicURL := fmt.Sprintf("http://%s/uploads/%s", serverIP, filename)

			// Persist uploaded image URL into request DTO so it is stored with the event
			body.PictureURL = &publicURL
		}

		// --- permission check ---
		if !canPostAs(viewerFrom(c), body.PostedAs.OrgPath, body.PostedAs.PositionKey) {
			return c.Status(fiber.StatusForbidden).
				JSON(dto.ErrorResponse{Error: "forbidden: you cannot post as this role"})
		}

		if body.NodeID == "" {
			return c.Status(fiber.StatusBadRequest).
				JSON(fiber.Map{"error": "node_id is required"})
		}
		if body.PostedAs == nil {
			return c.Status(fiber.StatusBadRequest).
				JSON(fiber.Map{"error": "posted_as is required"})
		}
		if body.Visibility == nil {
			body.Visibility = &models.Visibility{
				Access: "public",
			}
		} else if body.Visibility.Access == "" {
			body.Visibility.Access = "public"
		}

		// --- create event ---
		result, err := services.CreateEventWithSchedules(body, c.Context())
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		// Optionally append image URL in response
		if imgURL := body.PictureURL; imgURL != nil {
			return c.Status(fiber.StatusCreated).JSON(fiber.Map{
				"result":    result,
				"image_url": imgURL,
			})
		}

		return c.Status(fiber.StatusCreated).JSON(result)
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
		log.Printf("[DEBUG] viewerID=%s", viewerID.Hex())

		orgSets, err := services.AllUserOrg(viewerID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get user orgs"})
		}
		log.Printf("[DEBUG] orgSets=%+v", orgSets)

		events, err := services.GetVisibleEvents(viewerID, ctx, orgSets)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(events)
	}
}

// GetEventDetailHandler godoc
// @Summary Get individule event detail
// @Description Get full event detail including schedules and form ID (if any)
// @Tags events
// @Produce json
// @Param event_id path string true "Event ID"
// @Success 200 {object} dto.EventDetail
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /event/{eventId} [get]
func GetEventDetailHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		eventID := c.Params("event_id")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		event, err := services.GetEventDetail(eventID, ctx)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(event)
	}
}

// ParticipateEventWithNoFormHandler godoc
// @Summary Participate in an event (no form required)
// @Description Join an event directly if the event does not require a form submission
// @Tags events
// @Param event_id path string true "Event ID"
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /event/participate/{eventId} [post]
func ParticipateEventWithNoFormHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		eventID := c.Params("event_id")

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
// @Param event_id path string true "Event ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /event/{event_id} [delete]
func DeleteEventHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		collection_event := database.DB.Collection("events")

		idHex := c.Params("event_id")
		eventID, err := bson.ObjectIDFromHex(idHex)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, err = collection_event.UpdateOne(ctx, bson.M{"_id": eventID},
			bson.M{"$set": bson.M{"status": "inactive", "updated_at": time.Now()}})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete event"})
		}

		return c.JSON(fiber.Map{"message": "Event deleted (soft)"})
	}
}

// UpdateEventHandler godoc
// @Summary Update an event
// @Description Update any field of an event
// @Tags events
// @Accept json
// @Produce json
// @Param event_id path string true "Event ID"
// @Param body body map[string]interface{} true "Fields to update"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /event/{event_id} [patch]
// func UpdateEventHandler() fiber.Handler {
// 	return func(c *fiber.Ctx) error {
// 		eventIDHex := c.Params("event_id")
// 		eventID, err := bson.ObjectIDFromHex(eventIDHex)
// 		if err != nil {
// 			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid Event ID"})
// 		}

// 		var req dto.EventRequestDTO
// 		if err := c.BodyParser(&req); err != nil {
// 			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
// 		}

// 		file, err := c.FormFile("file")
// 		if err == nil && file != nil {
// 			timestamp := time.Now().UnixNano() / 1e6
// 			ext := filepath.Ext(file.Filename)
// 			filename := fmt.Sprintf("event_%d%s", timestamp, ext)
// 			savePath := filepath.Join("/var/www/html/uploads", filename)

// 			if err := c.SaveFile(file, savePath); err != nil {
// 				return c.Status(fiber.StatusInternalServerError).
// 					JSON(fiber.Map{"error": "failed to save file"})
// 			}

// 			publicURL := fmt.Sprintf("http://%s/uploads/%s", serverIP, filename)

// 			// Persist uploaded image URL into request DTO so it is stored with the event
// 			req.PictureURL = &publicURL
// 		}

// 		if req.NodeID == "" {
// 			return c.Status(fiber.StatusBadRequest).
// 				JSON(fiber.Map{"error": "node_id is required"})
// 		}
// 		if req.PostedAs == nil {
// 			return c.Status(fiber.StatusBadRequest).
// 				JSON(fiber.Map{"error": "posted_as is required"})
// 		}
// 		if req.Visibility == nil {
// 			req.Visibility = &models.Visibility{
// 				Access: "public",
// 			}
// 		} else if req.Visibility.Access == "" {
// 			req.Visibility.Access = "public"
// 		}

// 		if !canPostAs(viewerFrom(c), body.PostedAs.OrgPath, body.PostedAs.PositionKey) {
// 			return c.Status(fiber.StatusForbidden).
// 				JSON(dto.ErrorResponse{Error: "forbidden: you cannot post as this role"})
// 		}

// 		// Convert DTO to map for MongoDB $set
// 		update := bson.M{
// 			"node_id":           req.NodeID,
// 			"topic":             req.Topic,
// 			"description":       req.Description,
// 			"picture_url":       req.PictureURL,
// 			"max_participation": req.MaxParticipation,
// 			"posted_as":         req.PostedAs,
// 			"visibility":        req.Visibility,
// 			"org_of_content":    req.OrgOfContent,
// 			"status":            req.Status,
// 			"have_form":         req.Have_form,
// 			"schedules":         req.Schedules,
// 			"updated_at":        time.Now(),
// 		}

// 		collection := database.DB.Collection("event")
// 		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 		defer cancel()

// 		res, err := collection.UpdateOne(ctx,
// 			bson.M{"_id": eventID},
// 			bson.M{"$set": update},
// 		)
// 		if err != nil {
// 			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update event"})
// 		}
// 		if res.MatchedCount == 0 {
// 			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Event not found"})
// 		}

// 		return c.JSON(fiber.Map{"message": "Event updated successfully"})
// 	}
// }
// }
