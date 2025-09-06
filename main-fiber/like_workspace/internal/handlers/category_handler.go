package handlers

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"like_workspace/internal/repository"
)

// GetPostsWithCategories godoc
// @Summary Get posts with categories
// @Description ดึงโพสต์ทั้งหมดพร้อมหมวดหมู่ที่เกี่ยวข้อง
// @Tags posts
// @Produce json
// @Success 200 {array} repository.PostWithCategories
// @Failure 500 {object} map[string]interface{}
// @Router /api/posts/all-joined [get]

func GetPostsExistingInPC(client *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		items, err := repository.FetchPostsThatExistInPostCategories(c.Context(), client)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(items)
	}
}

func ToObjectID(hex string) (primitive.ObjectID, error) {
    hex = strings.TrimSpace(hex)
    oid, err := primitive.ObjectIDFromHex(hex)
    if err != nil {
        return primitive.NilObjectID, fmt.Errorf("invalid ObjectID: %w", err)
    }
    return oid, nil
}

func GetPostsWantedPC(client *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// รับจาก query param: /api/posts/wanted?id=66c6248b98c56c39f018e7d5
		idStr := strings.TrimSpace(c.Query("id"))
		if idStr == "" {
			return c.Status(400).JSON(fiber.Map{"error": "missing category id"})
		}

		// แปลง string → ObjectID
		oid, err := primitive.ObjectIDFromHex(idStr)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "invalid category id (must be 24-hex)",
				"id":    idStr,
			})
		}

		// เรียก repository
		items, err := repository.FetchPostsWantedCategories(c.Context(), client, oid)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(items)
	}
}
