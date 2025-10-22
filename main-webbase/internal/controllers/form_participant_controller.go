package controllers

import (
    "time"
    "context"

    "github.com/gofiber/fiber/v2"
    "go.mongodb.org/mongo-driver/v2/bson"
    "strings"
    "sort"

    "main-webbase/database"
    "main-webbase/dto"
    "main-webbase/internal/middleware"
    "main-webbase/internal/models"
    repo "main-webbase/internal/repository"
    "main-webbase/internal/services"
)

// InitializeFormHandler godoc
// @Summary Initialize a new form for an event
// @Description Create a new form associated with the specified event ID. Returns the created form details.
// @Tags forms
// @Accept json
// @Produce json
// @Param eventId path string true "Event ID to initialize form"
// @Success 200 {object} models.Event_form "Form initialized successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /event/{eventId}/form/initialize [post]
func InitializeFormHandler() fiber.Handler {
    return func(c *fiber.Ctx) error {
        eventID := c.Params("eventId")

		form, err := services.InitializeFormService(eventID, c.Context())
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{
			"message": "form initialized successfully",
			"data":    form,
		})
	}
}

// DisableFormHandler godoc
// @Summary Disable a form for an event
// @Description Marks a form as disabled for a specified event ID.
// @Tags forms
// @Accept json
// @Produce json
// @Param eventId path string true "Event ID to disable form"
// @Success 200 {object} map[string]string "Form disabled successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /event/{eventId}/form/disable [post]
func DisableFormHandler() fiber.Handler {
    return func(c *fiber.Ctx) error {
        eventID := c.Params("eventId")

		err := services.DisableFormService(eventID, c.Context())
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{
			"message": "Form Disable successfully",
		})
	}
}

// CreateFormQuestionHandler godoc
// @Summary Create or replace form questions
// @Description Replaces all questions for a given form with the provided question list. Requires permission to manage the event.
// @Tags forms
// @Accept json
// @Produce json
// @Param body body dto.FormQuestionCreateDTO true "Form questions payload"
// @Success 200 {array} models.Event_form_question "Questions replaced successfully"
// @Failure 400 {object} map[string]string "Invalid request or missing FormID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "No permission to manage this event"
// @Failure 404 {object} map[string]string "Form not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /event/{eventId}/form/questions [post]
func CreateFormQuestionHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body dto.FormQuestionCreateDTO
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		eventID := c.Params("eventId")
		if eventID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "eventId is required"})
		}

		form, err := repo.FindFormByEventID(c.Context(), eventID)
		if err != nil {
			return fiber.NewError(fiber.StatusNotFound, "form not found")
		}

		uid, err := middleware.UIDFromLocals(c)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
		}
		userPolicy, err := services.MyUserPolicy(c.Context(), uid)
		if err != nil {
			return fiber.NewError(fiber.StatusNotFound, "target policy not found")
		}

		if err := services.CanManageEvent(c.Context(), userPolicy, form.Event_ID.Hex()); err != nil {
			return fiber.NewError(fiber.StatusForbidden, "no permission to manage this event")
		}

		Questions_list, err := services.CreateFormQuestion(form.ID, body, c.Context())
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{
			"message": "questions replaced successfully",
			"data":    Questions_list, // Check permission
		})
	}
}

// GetFormQuestionHandler godoc
// @Summary Get form questions
// @Description Fetch all questions associated with a specific form ID.
// @Tags forms
// @Accept json
// @Produce json
// @Param formId path string true "Form ID"
// @Success 200 {object} map[string]interface{} "Questions fetched successfully"
// @Failure 400 {object} map[string]string "Form ID required"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /event/{eventId}/form/questions [get]
func GetFormQuestionHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		eventID := c.Params("eventId")
		if eventID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "eventId is required"})
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		questions, err := services.GetFormQuestion(ctx, eventID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{
			"message":   "questions fetched successfully",
			"Questions": questions,
		})
	}
}

// CreateUserAnswerHandler godoc
// @Summary Submit user answers for a form
// @Description Submit answers to a form. A user can only submit once per form.
// @Tags forms
// @Accept json
// @Produce json
// @Param body body dto.FormResponseSubmitDTO true "User answers payload"
// @Success 200 {object} map[string]interface{} "Response submitted successfully"
// @Failure 400 {object} map[string]string "Invalid request or user already submitted"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /event/{eventId}/form/answers [post]
func CreateUserAnswerHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body dto.FormResponseSubmitDTO
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		eventID := c.Params("eventId")
		if eventID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "eventId is required"})
		}

		form, err := repo.FindFormByEventID(c.Context(), eventID)
		if err != nil {
			return fiber.NewError(fiber.StatusNotFound, "form not found")
		}

		userID, err := middleware.UIDFromLocals(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
		}

		// Check if user send response already?
		exists, err := repo.HasUserSubmittedResponse(c.Context(), form.ID.Hex(), userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		if exists {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "User has already submitted a response for this form",
			})
		}

		response, err := services.SubmitFormResponse(form.ID, body, c.Context(), userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{
			"message": "response submitted successfully",
			"Answers": response,
		})
	}
}

// GetAllUserAnswerandQuestionHandler godoc
// @Summary Get all user answers with questions
// @Description Fetch all submitted responses for a given form along with the corresponding questions.
// @Tags forms
// @Accept json
// @Produce json
// @Param formId path string true "Form ID"
// @Success 200 {object} dto.FormMatrixResponseDTO "Form answers fetched successfully"
// @Failure 400 {object} map[string]string "Form ID required"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /event/{eventId}/form/matrix [get]
func GetAllUserAnswerandQuestionHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		eventID := c.Params("eventId")
		if eventID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "eventId is required"})
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		formmatrix, err := services.GetAllResponse(ctx, eventID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{
			"message": "form answers fetched successfully",
			"data":    formmatrix,
		})
	}
}

// UpdateParticipantStatusHandler godoc
// @Summary Update participant status
// @Description Update a participant's status (accept, stall, reject) for a specific event.
// @Tags participants
// @Accept json
// @Produce json
// @Param body body dto.UpdateParticipantStatusDTO true "Participant status payload"
// @Success 201 {object} map[string]string "Update user status success"
// @Failure 400 {object} map[string]string "Missing required fields"
// @Failure 500 {object} map[string]string "Failed to update status"
// @Router /event/participant/status [put]
func UpdateParticipantStatusHandler() fiber.Handler {
    return func(c *fiber.Ctx) error {
        var body dto.UpdateParticipantStatusDTO
        if err := c.BodyParser(&body); err != nil {
            return fiber.NewError(fiber.StatusBadRequest, "invalid body")
        }

        if body.UserID == "" || body.EventID == "" || body.Status == "" {
            return fiber.NewError(fiber.StatusBadRequest, "user_id, event_id, and status are required")
        }

        // Require organizer permission for this event
        uid, err := middleware.UIDFromLocals(c)
        if err != nil {
            return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
        }
        userObjID, err := bson.ObjectIDFromHex(uid)
        if err != nil {
            return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
        }
        eventObjID, err := bson.ObjectIDFromHex(body.EventID)
        if err != nil {
            return fiber.NewError(fiber.StatusBadRequest, "invalid event_id")
        }
        // Must be organizer of this event
        if err := database.DB.Collection("event_participant").FindOne(c.Context(), bson.M{
            "event_id": eventObjID,
            "user_id":  userObjID,
            "role":     "organizer",
        }).Err(); err != nil {
            return fiber.NewError(fiber.StatusForbidden, "no permission to manage this event")
        }

        if err := services.UpdateParticipantStatus(c.Context(), body); err != nil {
            return fiber.NewError(fiber.StatusInternalServerError, "failed to update user status")
        }

        // parameter for notification
        ref := models.Ref{
            ID: eventObjID,
            Entity: "event",
        }
        
        colEvent := database.DB.Collection("events")
        colNoti := database.DB.Collection("notification")
        var result struct {
            Title string `bson:"topic"`
        }
        err = colEvent.FindOne(c.Context(), bson.M{"_id": eventObjID}).Decode(&result)
        if err != nil {
            return fiber.NewError(fiber.StatusInternalServerError, "failed to fetch event title: " + err.Error())
        }

        notiParam := models.NotiParams{
            EventTitle: result.Title,
            EventID: eventObjID,
        }
        if err := services.NotifyOne(c.Context(), 
            colNoti, 
            userObjID, 
            services.NotiAuditionApproved, 
            ref,
            notiParam); err != nil { 
            return fiber.NewError(fiber.StatusInternalServerError, "failed to send notification")
            }

        return c.Status(fiber.StatusCreated).JSON(fiber.Map{
            "message": "Update User Status Success",
            "data":    body.Status,
        })
    }
}

// ListManagedEventsHandler
// GET /event/managed
// Returns events where the current user is an organizer, with counts of pending/accepted participants.
func ListManagedEventsHandler() fiber.Handler {
    return func(c *fiber.Ctx) error {
        uid, err := middleware.UIDFromLocals(c)
        if err != nil {
            return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
        }
        userID, err := bson.ObjectIDFromHex(uid)
        if err != nil {
            return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
        }

        colEP := database.DB.Collection("event_participant")
        cur, err := colEP.Find(c.Context(), bson.M{"user_id": userID, "role": "organizer"})
        if err != nil {
            return fiber.NewError(fiber.StatusInternalServerError, err.Error())
        }
        defer cur.Close(c.Context())

        type OrgRow struct { EventID bson.ObjectID `bson:"event_id"` }
        seen := make(map[bson.ObjectID]struct{})
        eventIDs := make([]bson.ObjectID, 0, 8)
        for cur.Next(c.Context()) {
            var r struct{ EventID bson.ObjectID `bson:"event_id"` }
            if err := cur.Decode(&r); err == nil {
                if _, ok := seen[r.EventID]; !ok {
                    seen[r.EventID] = struct{}{}
                    eventIDs = append(eventIDs, r.EventID)
                }
            }
        }
        if len(eventIDs) == 0 {
            return c.JSON([]fiber.Map{})
        }

        // For each event, fetch event info + counts
        out := make([]fiber.Map, 0, len(eventIDs))
        for _, eid := range eventIDs {
            ev, _ := repo.GetEventByID(c.Context(), eid)
            topic := ""
            maxPart := 0
            if ev != nil {
                topic = ev.Topic
                maxPart = ev.MaxParticipation
            }

            // Count participants by status
            stallCnt, _ := database.DB.Collection("event_participant").CountDocuments(c.Context(), bson.M{"event_id": eid, "role": "participant", "status": "stall"})
            acceptCnt, _ := database.DB.Collection("event_participant").CountDocuments(c.Context(), bson.M{"event_id": eid, "role": "participant", "status": "accept"})

            out = append(out, fiber.Map{
                "eventId":           eid.Hex(),
                "topic":             topic,
                "max_participation": maxPart,
                "pendingCount":      stallCnt,
                "acceptedCount":     acceptCnt,
            })
        }
        return c.JSON(out)
    }
}

// ListEventParticipantsHandler
// GET /event/:eventId/participants?status=&role=
// Returns participants for an event, with basic user info.
func ListEventParticipantsHandler() fiber.Handler {
    return func(c *fiber.Ctx) error {
        eventIDHex := c.Params("eventId")
        eid, err := bson.ObjectIDFromHex(eventIDHex)
        if err != nil { return fiber.NewError(fiber.StatusBadRequest, "invalid eventId") }

        // Permission: must be organizer
        uid, err := middleware.UIDFromLocals(c)
        if err != nil { return fiber.NewError(fiber.StatusUnauthorized, "unauthorized") }
        uidObj, err := bson.ObjectIDFromHex(uid)
        if err != nil { return fiber.NewError(fiber.StatusUnauthorized, "unauthorized") }
        if err := database.DB.Collection("event_participant").FindOne(c.Context(), bson.M{
            "event_id": eid,
            "user_id":  uidObj,
            "role":     "organizer",
        }).Err(); err != nil {
            return fiber.NewError(fiber.StatusForbidden, "no permission to manage this event")
        }

        role := c.Query("role")
        if role == "" { role = "participant" }
        status := c.Query("status") // optional: accept|stall|reject

        filter := bson.M{"event_id": eid}
        if role != "" { filter["role"] = role }
        if status != "" { filter["status"] = status }

        cur, err := database.DB.Collection("event_participant").Find(c.Context(), filter)
        if err != nil { return fiber.NewError(fiber.StatusInternalServerError, err.Error()) }
        defer cur.Close(c.Context())

        // Build response with user info
        out := make([]fiber.Map, 0, 20)
        colUsers := database.DB.Collection("users")
        for cur.Next(c.Context()) {
            var p models.Event_participant
            if err := cur.Decode(&p); err != nil { continue }
            var user struct{ ID bson.ObjectID `bson:"_id"`; First string `bson:"firstname"`; Last string `bson:"lastname"` }
            _ = colUsers.FindOne(c.Context(), bson.M{"_id": p.User_ID}).Decode(&user)
            out = append(out, fiber.Map{
                "user_id":  p.User_ID.Hex(),
                "first_name": user.First,
                "last_name":  user.Last,
                "role":     p.Role,
                "status":   p.Status,
                "response_id": func() string { if p.Response_ID.IsZero() { return "" } else { return p.Response_ID.Hex() } }(),
            })
        }
        return c.JSON(out)
    }
}

// ManageableOrgsHandler
// GET /event/manageable-orgs?search=
// Returns org nodes for which the current user has create permission (organize:create or event:create).
func ManageableOrgsHandler() fiber.Handler {
    return func(c *fiber.Ctx) error {
        uid, err := middleware.UIDFromLocals(c)
        if err != nil { return fiber.NewError(fiber.StatusUnauthorized, "unauthorized") }

        policies, err := services.MyUserPolicy(c.Context(), uid)
        if err != nil { return fiber.NewError(fiber.StatusInternalServerError, err.Error()) }

        // Collect org prefixes with create permission
        prefixes := make(map[string]struct{})
        for _, p := range policies {
            if !p.Enabled { continue }
            hasCreate := false
            for _, a := range p.Actions {
                if a == "organize:create" || a == "event:create" {
                    hasCreate = true
                    break
                }
            }
            if hasCreate && p.OrgPrefix != "" {
                prefixes[p.OrgPrefix] = struct{}{}
            }
        }

        // Build output with names
        q := strings.ToLower(strings.TrimSpace(c.Query("search")))
        out := make([]fiber.Map, 0, len(prefixes))
        for path := range prefixes {
            node, _ := repo.FindByOrgPath(c.Context(), path)
            if node == nil { continue }
            if q != "" {
                if !strings.Contains(strings.ToLower(node.Name), q) &&
                   !strings.Contains(strings.ToLower(node.ShortName), q) &&
                   !strings.Contains(strings.ToLower(node.OrgPath), q) {
                    continue
                }
            }
            out = append(out, fiber.Map{
                "org_path":   node.OrgPath,
                "name":       node.Name,
                "short_name": node.ShortName,
                "type":       node.Type,
                "node_id":    node.ID.Hex(),
            })
        }
        sort.Slice(out, func(i, j int) bool {
            si, sj := out[i]["short_name"], out[j]["short_name"]
            nis := strings.ToLower(strings.TrimSpace(func(v any) string { if v==nil {return ""}; return v.(string) }(si)))
            njs := strings.ToLower(strings.TrimSpace(func(v any) string { if v==nil {return ""}; return v.(string) }(sj)))
            return nis < njs
        })
        return c.JSON(out)
    }
}

// GetMyParticipantStatusHandler godoc
// @Summary Get current user's participant status
// @Description Returns the status of the authenticated user for a specific event.
// @Tags participants
// @Accept json
// @Produce json
// @Param eventId path string true "Event ID"
// @Success 200 {object} map[string]string "Current user status"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /event/participant/mystatus/{eventId} [get]
func GetMyParticipantStatusHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		eventID := c.Params("eventId")

		userID, err := middleware.UIDFromLocals(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
		}

		userStatus, err := services.GetParticipantStatus(c.Context(), userID, eventID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": userStatus,
		})
	}
}
