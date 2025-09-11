package handlers

import (
	"time"
	"context"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"like_workspace/dto"
	"like_workspace/services"
)

func CreateLikeHandler(client *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body dto.LikeRequestDTO
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid body"})
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		status, payload := services.Like(ctx, client, body)
		return c.Status(status).JSON(payload)
	}
}
