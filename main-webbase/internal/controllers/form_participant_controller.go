package controllers

import (
	// "context"
	// "time"

	"github.com/gofiber/fiber/v2"
	// "go.mongodb.org/mongo-driver/v2/bson"

	// "main-webbase/database"
	"main-webbase/dto"
	// "main-webbase/internal/models"
	"main-webbase/internal/services"
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
			"data":    Questions_list,
		})
	}
}

// func GetFormQuestionHandler() fiber.Handler {
// 	return func(c *fiber.Ctx) error {

// 	}
// }

// func CreateUserAnswerHandler() fiber.Handler {
// 	return func(c *fiber.Ctx) error {

// 	}
// }

// func GetAllUserAnswerandQuestionHandler() fiber.Handler {
// 	return func(c *fiber.Ctx) error {

// 	}
// }

// func CreateUserParticipantHandler() fiber.Handler {
// 	return func(c *fiber.Ctx) error {

// 	}
// }

// func GetAllParticipantStatusHandler() fiber.Handler {
// 	return func(c *fiber.Ctx) error {

// 	}
// }


// func GetMyParticipantStatusHandler() fiber.Handler {
// 	return func(c *fiber.Ctx) error {

// 	}
// }
