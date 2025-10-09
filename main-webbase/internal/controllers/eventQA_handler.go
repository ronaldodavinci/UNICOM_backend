package controllers

import (
	"time"
	"errors"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"main-webbase/dto"
	m "main-webbase/internal/middleware"
	"main-webbase/internal/models"
	repo "main-webbase/internal/repository"
	u "main-webbase/utils"
)

func CreateEventQAHandler(client *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		db := client.Database("unicom")
		colEvents := db.Collection("events")
		colEventQA := db.Collection("event_qa")

		// user id
		userID, err := m.UIDFromLocals(c)
		if err != nil {
			return err
		}
		uid, err := u.Oid(userID)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid user_id")
		}

		// event id
		eventID, err := u.Oid(c.Params("eventId"))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid eventId")
		}

		// body
		var req dto.CreateQADTO
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid body")
		}
		if len(req.QuestionText) == 0 {
			return fiber.NewError(fiber.StatusBadRequest, "questionText required")
		}

		// owner (answerer)
		answererID, err := repo.FindEventOrganizerID(colEvents, eventID, c.Context())
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return fiber.NewError(fiber.StatusNotFound, "event not found")
			}
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		doc := models.EventQA{
			ID:                bson.NewObjectID(),
			EventID:           eventID,
			QuestionerID:      uid,
			AnswererID:        answererID,
			QuestionText:      req.QuestionText,
			QuestionCreatedAt: time.Now().UTC(),
			Status:            "pending",
		}

		if _, err := colEventQA.InsertOne(c.Context(), doc); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		return c.Status(fiber.StatusCreated).JSON(models.EventQA(doc))
	}
}