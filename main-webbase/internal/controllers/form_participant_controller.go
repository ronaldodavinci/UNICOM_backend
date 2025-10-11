package controllers

import (
	"time"
	"context"

	"github.com/gofiber/fiber/v2"

	"main-webbase/dto"
	"main-webbase/internal/middleware"
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
		eventID := c.Params("value")

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
		eventID := c.Params("value")

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

		// uid, err := middleware.UIDFromLocals(c)
		// if err != nil {
		//     return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
		// }
		// userPolicy, err := services.MyUserPolicy(c.Context(), uid)
		// if err != nil {
		//     return fiber.NewError(fiber.StatusNotFound, "target policy not found")
		// }

		// if err := services.CanManageEvent(c.Context(), userPolicy, body.EventID); err != nil {
		// 	return fiber.NewError(fiber.StatusForbidden, "no permission to manage this event")
		// }

		if err := services.UpdateParticipantStatus(c.Context(), body); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to update user status")
		}
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"message": "Update User Status Success",
			"data":    body.Status,
		})
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
