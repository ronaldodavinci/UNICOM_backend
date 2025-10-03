package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"like_workspace/dto"
	"like_workspace/internal/accessctx"
	"like_workspace/internal/repository"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// GET /api/posts/category/any/cursor?limit=20&cursor=...&type=all|public|private
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
			visibility = "" // รวม public + private ที่ผู้ดูมีสิทธิ์
		case "public", "private":
			visibility = t
		default:
			return fiber.NewError(fiber.StatusBadRequest,
				fmt.Sprintf("invalid type %q (use all|public|private)", t))
		}

		curStr := c.Query("cursor")

		// ⬇️ ดึงสิทธิ์ของผู้ดูจาก Locals (ต้องมี middleware InjectViewer วางก่อนแล้ว)
		v, ok := c.Locals("viewer").(*accessctx.ViewerAccess)
		if !ok || v == nil {
			return fiber.ErrUnauthorized
		}

		items, next, total, err := repository.ListAllPostsVisibleToViewer(
			c.Context(), client, curStr, visibility, limit, v.SubtreeNodeIDs, // ⬅️ ส่งสิทธิ์ของผู้ดู
		)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		resp := dto.ListByCategoryResp[bson.M]{
			Items:      items,
			NextCursor: next,
			HasMore:    next != nil,
			TotalCount: total,
		}
		return c.JSON(resp)
	}
}
