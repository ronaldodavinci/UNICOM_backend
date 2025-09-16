package handlers

import (
	"net/http"
	"strings"
	"fmt"

	"like_workspace/dto"
	"like_workspace/internal/repository"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// GET /api/posts/category/any/cursor?limit=20&cursor=...
func GetPostsVisibilityCursor(client *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		limit := int64(c.QueryInt("limit", 10))
		if limit <= 0 {
			limit = 10
		}
		if limit > 20 {
			limit = 20
		}

		t := strings.ToLower(strings.TrimSpace(c.Query("type")))
		var visibility string
		switch t {
		case "", "all":
			visibility = "" // ไม่กรอง
		case "public", "private":
			visibility = t
		default:
			return fiber.NewError(fiber.StatusBadRequest,
				fmt.Sprintf("invalid type %q (use all|public|private)", t))
		}

		curStr := c.Query("cursor")

		items, next, total, err := repository.ListAllPostsWithVisibilityNewestFirst(c.Context(), client, curStr, visibility, limit)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		// ใช้ชนิดให้ตรงกับผลลัพธ์ []bson.M เพื่อเลี่ยง type mismatch
		resp := dto.ListByCategoryResp[bson.M]{
			Items:      items,
			NextCursor: next,
			HasMore:    next != nil,
			TotalCount: total,
		}
		return c.JSON(resp)
	}
}