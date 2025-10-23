package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"time"
	"strings"	

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"

	"main-webbase/database"
	"main-webbase/dto"
	"main-webbase/internal/middleware"
	"main-webbase/internal/models"
	"main-webbase/internal/services"
)

// CreateEventHandler godoc
// @Summary Create new event
// @Description Create an event with optional image upload
// @Tags events
// @Accept multipart/form-data
// @Produce json
// @Param file formData file false "Upload event image"
// @Param NodeID formData string true "Node ID"
// @Param postedAs.org_path formData string true "Organization path"
// @Param postedAs.position_key formData string true "Position key"
// @Success 201 {object} dto.EventCreateResult
// @Failure 400 {object} map[string]string
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} map[string]string
// @Router /event [post]
func CreateEventHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body dto.EventRequestDTO

		// Initialize nested pointers first to avoid nil deref
		if body.PostedAs == nil {
			body.PostedAs = &models.PostedAs{}
		}

		// Parse minimal required form fields
		body.NodeID = c.FormValue("NodeID")
		body.PostedAs.OrgPath = c.FormValue("postedAs.org_path")
		body.PostedAs.PositionKey = c.FormValue("postedAs.position_key")

		// Optional basic fields
		if v := c.FormValue("topic"); v != "" {
			body.Topic = v
		}
		if v := c.FormValue("description"); v != "" {
			body.Description = v
		}
		if v := c.FormValue("org_of_content"); v != "" {
			body.OrgOfContent = v
		}
		if v := c.FormValue("status"); v != "" {
			body.Status = v
		}
		if v := c.FormValue("max_participation"); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				body.MaxParticipation = n
			}
		}
		if v := c.FormValue("have_form"); v != "" {
			switch v {
			case "1", "true", "TRUE", "True", "yes", "YES", "Yes":
				body.Have_form = true
			default:
				body.Have_form = false
			}
		}

		// Visibility (JSON string in multipart form)
		if v := c.FormValue("visibility"); v != "" {
			var vis models.Visibility
			if err := json.Unmarshal([]byte(v), &vis); err == nil {
				body.Visibility = &vis
			}
		}
		// Safe default to avoid nil visibility panics in service layer
		if body.Visibility == nil {
			body.Visibility = &models.Visibility{Access: "public"}
		}

		if body.NodeID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "NodeID is required"})
		}
		if body.PostedAs.OrgPath == "" || body.PostedAs.PositionKey == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "postedAs.org_path and postedAs.position_key are required"})
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
			body.PictureURL = publicURL
			// Also keep in locals for convenience in response
			c.Locals("event_image_url", publicURL)
		}

		// --- permission check ---
		if !canPostAs(viewerFrom(c), body.PostedAs.OrgPath, body.PostedAs.PositionKey) {
			return c.Status(fiber.StatusForbidden).
				JSON(dto.ErrorResponse{Error: "forbidden: you cannot post as this role"})
		}

		// --- create event ---
		result, err := services.CreateEventWithSchedules(body, c.Context())
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		// Optionally append image URL in response
		if imgURL := c.Locals("event_image_url"); imgURL != nil {
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

		// ---------- parse filters ----------
        q := c.Query("q")
		roles := splitCSVFilter(c.Query("role"))

		// events, err := services.GetVisibleEvents(viewerID, ctx, orgSets)
		events, err := services.GetVisibleEventsFiltered(
            viewerID,
            ctx,
            orgSets,
            services.VisibleEventQuery{
                Roles:    roles,
                Q:        q,
            },
        )

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(events)
	}
}

func splitCSVFilter(s string) []string {
	if s == "" {
		return []string{} // อย่าคืน nil ถ้าจะใช้กับ $in
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
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
// @Router /event/{id} [delete]
func DeleteEventHandler(c *fiber.Ctx) error {
	collection_event := database.DB.Collection("event")

	idHex := c.Params("event_id")
	event_id, err := bson.ObjectIDFromHex(idHex)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = collection_event.UpdateOne(ctx, bson.M{"_id": event_id},
		bson.M{"$set": bson.M{"status": "inactive", "updated_at": time.Now()}})
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
