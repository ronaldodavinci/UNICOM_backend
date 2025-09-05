package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/mongo"

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

func GetPostsByCategoriesCursor(client *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// query: ?limit=5&cursor=xxx
		limit := int64(5)
		if s := c.Query("limit"); s != "" {
			if v, err := strconv.ParseInt(s, 10, 64); err == nil && v > 0 {
				limit = v
			}
		}
		cursor := c.Query("cursor")

		items, next, err := repository.FetchPostsWithCategoriesCursor(c.Context(), client, limit, cursor, nil, nil)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{
			"items":       items,
			"next_cursor": next, // ว่างถ้าไม่มีหน้าถัดไป
		})
	}
}

func GetPostsByCategory(client *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		catID := c.Params("categoryID")
		limitStr := c.Query("limit", "10")
		limit, _ := strconv.ParseInt(limitStr, 10, 64)

		items, err := repository.FetchPostsByCategory(c.Context(), client, catID, limit)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(items)
	}
}