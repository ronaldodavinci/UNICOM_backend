package controllers

import (
    "time"
    "errors"

    "github.com/gofiber/fiber/v2"
    "go.mongodb.org/mongo-driver/v2/bson"
    "go.mongodb.org/mongo-driver/v2/mongo"
    "go.mongodb.org/mongo-driver/v2/mongo/options"

    "main-webbase/dto"
    "main-webbase/database"
    m "main-webbase/internal/middleware"
    "main-webbase/internal/models"
    repo "main-webbase/internal/repository"
    s "main-webbase/internal/services"
)

func CreateEventQAHandler(client *mongo.Client) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Use configured DB (same as rest of the app)
        colEvents := database.DB.Collection("events")
        colEventQA := database.DB.Collection("event_qa")

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
        colEventQA := database.DB.Collection("event_qa")
        colParticipant := database.DB.Collection("event_participant")

        // 1) current user (must be organizer of the event)
        uidStr, err := m.UIDFromLocals(c)
        if err != nil {
            return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
        }
        uid, err := bson.ObjectIDFromHex(uidStr)
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

        // 4) load QA to get event_id & ensure still pending
        var qa models.EventQA
        if err := colEventQA.FindOne(c.Context(), bson.M{"_id": qaID, "status": "pending"}).Decode(&qa); err != nil {
            return fiber.NewError(fiber.StatusForbidden, "cannot answer")
        }

        // 5) permission: must be organizer of this event
        if err := colParticipant.FindOne(c.Context(), bson.M{
            "event_id": qa.EventID,
            "user_id":  uid,
            "role":     "organizer",
        }).Err(); err != nil {
            return fiber.NewError(fiber.StatusForbidden, "cannot answer")
        }

        // 6) update answer
        res := colEventQA.FindOneAndUpdate(
            c.Context(),
            bson.M{"_id": qaID, "status": "pending", "answer_text": bson.M{"$eq": nil}},
            bson.M{"$set": bson.M{
                "answer_text":       req.AnswerText,
                "answer_created_at": time.Now().UTC(),
                "answerer_id":       uid,
                "status":            "answered",
            }},
            options.FindOneAndUpdate().SetReturnDocument(options.After),
        )

        var out models.EventQA
        if err := res.Decode(&out); err != nil {
            return fiber.NewError(fiber.StatusForbidden, "cannot answer")
        }

        // parameter for notification
        ref := models.Ref{
            ID: qaID,
            Entity: "qa",
        }
        
        colEvent := database.DB.Collection("events")
        colNoti := database.DB.Collection("notification")
        var result struct {
            Title string `bson:"topic"`
        }
        err = colEvent.FindOne(c.Context(), bson.M{"_id": qa.EventID}).Decode(&result)
        if err != nil {
            return fiber.NewError(fiber.StatusInternalServerError, "failed to fetch event title: " + err.Error())
        }

        notiParam := models.NotiParams{
            EventTitle: result.Title,
            EventID: qa.EventID,
        }
        if err := s.NotifyOne(c.Context(), 
            colNoti, 
            uid, 
            s.NotiQAAnswered, 
            ref,
            notiParam); err != nil { 
            return fiber.NewError(fiber.StatusInternalServerError, "failed to send notification")
            }

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

// ListEventQAHandler lists Q&A for a given event
func ListEventQAHandler(client *mongo.Client) fiber.Handler {
    return func(c *fiber.Ctx) error {
        eventIDHex := c.Params("eventId")
        eid, err := bson.ObjectIDFromHex(eventIDHex)
        if err != nil {
            return fiber.NewError(fiber.StatusBadRequest, "invalid eventId")
        }

        col := database.DB.Collection("event_qa")
        cur, err := col.Find(c.Context(), bson.M{"event_id": eid, "status": bson.M{"$ne": "deleted"}}, options.Find().SetSort(bson.D{{Key: "question_created_at", Value: -1}}))
        if err != nil {
            return fiber.NewError(fiber.StatusInternalServerError, err.Error())
        }
        defer cur.Close(c.Context())

        out := make([]dto.EventQAResponse, 0, 20)
        for cur.Next(c.Context()) {
            var q models.EventQA
            if err := cur.Decode(&q); err != nil { continue }
            out = append(out, dto.EventQAResponse{
                ID:                q.ID.Hex(),
                EventID:           q.EventID.Hex(),
                QuestionerID:      q.QuestionerID.Hex(),
                AnswererID:        q.AnswererID.Hex(),
                QuestionText:      q.QuestionText,
                QuestionCreatedAt: q.QuestionCreatedAt,
                AnswerText:        q.AnswerText,
                AnswerCreatedAt:   q.AnswerCreatedAt,
                Status:            q.Status,
            })
        }

        return c.JSON(out)
    }
}
