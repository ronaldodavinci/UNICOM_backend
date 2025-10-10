package controllers

import (
	"time"
	"errors"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"main-webbase/dto"
	m "main-webbase/internal/middleware"
	"main-webbase/internal/models"
	repo "main-webbase/internal/repository"
)

func CreateEventQAHandler(client *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		db := client.Database("unicom")
		colEvents := db.Collection("events")
		colEventQA := db.Collection("event_qa")

		// user id
		QuestionerID, err := m.UIDFromLocals(c)
		if err != nil {
			return err
		}
		uid, err := bson.ObjectIDFromHex(QuestionerID)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid user_id")
		}

		// event id
		eventID, err := bson.ObjectIDFromHex(c.Params("eventId"))
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

func AnswerEventQAHandler(client *mongo.Client) fiber.Handler {
    return func(c *fiber.Ctx) error {
		db := client.Database("unicom")
		colEventQA := db.Collection("event_qa")

        // 1) answerer (เจ้าของ event)
        answererID, err := m.UIDFromLocals(c)
        if err != nil {
            return fiber.NewError(fiber.StatusBadRequest, "invalid user_id")
        }

        // 2) qaId
        qaIDHex := c.Params("qaId")
        qaID, err := bson.ObjectIDFromHex(qaIDHex)
        if err != nil {
            return fiber.NewError(fiber.StatusBadRequest, "invalid qaId")
        }

        // 3) body
        var req dto.AnswerQADTO
        if err := c.BodyParser(&req); err != nil {
            return fiber.NewError(fiber.StatusBadRequest, "invalid body")
        }
        if len(req.AnswerText) == 0 {
            return fiber.NewError(fiber.StatusBadRequest, "answerText required")
        }

        // 4) update เฉพาะกรณียังไม่ได้ตอบ และผู้ตอบเป็นเจ้าของ
        filter := bson.M{
            "_id":         qaID,
            "status":      "pending",
            "answer_text": bson.M{"$eq": nil},
            "answerer_id": answererID,
        }
        update := bson.M{"$set": bson.M{
            "answer_text":       req.AnswerText,
            "answer_created_at": time.Now().UTC(),
            "status":            "answered",
        }}

        res := colEventQA.FindOneAndUpdate(
            c.Context(),
            filter,
            update,
            options.FindOneAndUpdate().SetReturnDocument(options.After),
        )

        var out models.EventQA
        if err := res.Decode(&out); err != nil {
            // ไม่เจอ/ไม่ใช่เจ้าของ/ตอบไปแล้ว → บอกให้เข้าใจง่าย
            return fiber.NewError(fiber.StatusForbidden, "cannot answer")
        }

        // 5) response
        return c.JSON(dto.EventQAResponse{
            ID:                out.ID.Hex(),
            EventID:           out.EventID.Hex(),
            QuestionerID:      out.QuestionerID.Hex(),
            AnswererID:        out.AnswererID.Hex(),
            QuestionText:      out.QuestionText,
            QuestionCreatedAt: out.QuestionCreatedAt,
            AnswerText:        out.AnswerText,
            AnswerCreatedAt:   out.AnswerCreatedAt,
            Status:            out.Status,
        })
    }
}
