package handlers

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/pllus/main-fiber/lukkate_workspace/internal/db"
)

func GetUsers(c *fiber.Ctx) error {
	if db.Client == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Mongo client is not initialized",
			"data":    nil,
		})
	}

	collection := db.Client.Database("User_1").Collection("User")

	cursor, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to query database",
			"data":    nil,
		})
	}
	defer cursor.Close(context.Background())

	var users []bson.M
	if err := cursor.All(context.Background(), &users); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to decode documents",
			"data":    nil,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Users fetched successfully",
		"data":    users,
	})
}
