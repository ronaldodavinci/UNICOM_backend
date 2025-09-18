package handlers

import (
	"time"
	"context"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/bson"

	"like_workspace/dto"
	"like_workspace/services"
)

func fetchUserID(c *fiber.Ctx) (bson.ObjectID, bool) {
	if v := c.Locals("user_id"); v != nil {
		if s, ok := v.(string); ok {
			if oid, err := bson.ObjectIDFromHex(s); err == nil {
				return oid, true
			}
		}
	}
	return bson.NilObjectID, false
}



func LikeUnlikeHandler(client *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user_id, ok := fetchUserID(c) // <- ใช้ฟังก์ชันเดิมของเพื่อน
		if !ok {
			return c.Status(fiber.StatusUnauthorized).
				JSON(dto.ErrorResponse{Message: "missing userId in context"})
		}

		var body dto.LikeRequestDTO
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid body"})
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		status, payload := services.Like(ctx, client, body, user_id)
		return c.Status(status).JSON(payload)
	}
}





