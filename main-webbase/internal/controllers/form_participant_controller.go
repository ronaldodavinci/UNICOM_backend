package controllers

import (
	// "context"
	"time"

	"context"

	"github.com/gofiber/fiber/v2"
	// "go.mongodb.org/mongo-driver/v2/bson"

	// "main-webbase/database"
	"main-webbase/dto"
	// "main-webbase/internal/models"
	"main-webbase/internal/services"
	"main-webbase/internal/middleware"
	repo "main-webbase/internal/repository"
)

func InitializeFormHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body dto.FormCreateDTO
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		if body.EventID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "NodeID is required"})
		}

		form, err := services.InitializeFormService(body, c.Context())
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{
			"message": "form initialized successfully",
			"data":    form,
		})
	}
}

// func DisableFormHandler() fiber.Handler {
// 	return func(c *fiber.Ctx) error {

// 	}
// }

func CreateFormQuestionHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body dto.FormQuestionCreateDTO
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		if body.FormID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "FormID is required"})
		}

		Questions_list, err := services.CreateFormQuestion(body, c.Context())
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{
			"message": "questions replaced successfully",
			"data":    Questions_list, // Check permission
		})
	}
}

func GetFormQuestionHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		formID := c.Params("form_id")
		if formID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "form_id is required"})
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		questions, err := services.GetFormQuestion(ctx, formID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{
			"message": "questions fetched successfully",
			"Questions": questions,
		})
	}
}

func CreateUserAnswerHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body dto.FormResponseSubmitDTO
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		if body.FormID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "FormID is required"})
		}

		userID, err := middleware.UIDFromLocals(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
		}

		// Check if user send response already?
		exists, err := repo.HasUserSubmittedResponse(c.Context(), body.FormID, userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		if exists {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "User has already submitted a response for this form",
			})
		}

		response, err := services.SubmitFormResponse(body, c.Context(), userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{
			"message": "response submitted successfully",
			"Answers": response,
		})
	}
}

func GetAllUserAnswerandQuestionHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		formID := c.Params("form_id")
		if formID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "form_id is required"})
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		formmatrix, err := services.GetAllResponse(ctx, formID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{
			"message": "form answers fetched successfully",
			"data": formmatrix,
		})
	}
}

// func UpdateParticipantStatusHandler() fiber.Handler {
// 	return func(c *fiber.Ctx) error {
// 		var body dto.UpdateParticipantStatusDTO
// 		if err := c.BodyParser(&body); err != nil {
//             return fiber.NewError(fiber.StatusBadRequest, "invalid body")
//         }

// 		// uid, err := middleware.UIDFromLocals(c)
//         // if err != nil {
//         //     return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
//         // }
// 		// userPolicy, err := services.MyUserPolicy(c.Context(), uid)
//         // if err != nil {
//         //     return fiber.NewError(fiber.StatusNotFound, "target policy not found")
//         // }

// 		// if err := services.CanManageEvent(c.Context(), userPolicy, body.EventID); err != nil {
// 		// 	return fiber.NewError(fiber.StatusForbidden, "no permission to manage this event")
// 		// }

// 		if err := services.UpdateParticipantStatus(c.Context(), body); err != nil {
// 			return fiber.NewError(fiber.StatusInternalServerError, "failed to update user status")
// 		}
// 		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
// 			"message": "Update User Status Success",
// 			"data": body.Status,
// 		})
// 	}
// }



// func GetMyParticipantStatusHandler() fiber.Handler {
// 	return func(c *fiber.Ctx) error {

// 	}
// }
