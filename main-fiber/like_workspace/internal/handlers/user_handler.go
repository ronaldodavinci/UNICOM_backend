package handlers

import (
	//"context"
	//"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"like_workspace/internal/repository"
	//"like_workspace/database"
	//"like_workspace/model"
)

// GetUserHandler godoc
// @Summary Get all users
// @Description Get all users from MongoDB User collection
// @Tags users
// @Accept  json
// @Produce  json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /Post [get]

func GetUserHandler(c *fiber.Ctx, client *mongo.Client) error {
	users, err := repository.FetchUsers(client)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
			"data":    nil,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Users fetched successfully",
		"data":    users,
	})
}

func GetPostsLimit(client *mongo.Client) fiber.Handler {
    return func(c *fiber.Ctx) error {
        const limit int64 = 5 // 🔒 locked limit

        collection := client.Database("test").Collection("Posts")
        opts := options.Find().SetLimit(limit).SetSort(bson.M{"timestamp": -1})

        cursor, err := collection.Find(c.Context(), bson.M{}, opts)
        if err != nil {
            return c.Status(500).SendString(err.Error())
        }
        defer cursor.Close(c.Context())

        var results []map[string]interface{}
        if err := cursor.All(c.Context(), &results); err != nil {
            return c.Status(500).SendString(err.Error())
        }

        return c.JSON(results)
    }
}
