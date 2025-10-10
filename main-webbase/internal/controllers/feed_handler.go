package controllers

import (
	"net/http"

	"main-webbase/dto"
	"main-webbase/internal/accessctx"
	"main-webbase/internal/repository"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// GetPostsVisibilityCursor godoc
// @Summary      Get feed posts with visibility & cursor pagination
// @Description  ดึงรายการโพสต์ที่ผู้ใช้มีสิทธิ์เห็น พร้อม cursor-based pagination (ตรวจสิทธิ์จาก JWT -> viewer Locals)
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        limit   query     int     false  "จำนวน item ต่อหน้า (1–20)"  minimum(1) maximum(20) default(10)
// @Param        cursor  query     string  false  "Cursor จากหน้าก่อนหน้า (base64)"
// @Success      200     {object}  dto.ListByCategoryResp
// @Security     BearerAuth
// @Router       /posts [get]
func GetPostsVisibilityCursor(client *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		limit := int64(c.QueryInt("limit", 10))
		if limit <= 0 {
			limit = 10
		}
		if limit > 20 {
			limit = 20
		}

		curStr := c.Query("cursor")

		// ⬇️ ดึงสิทธิ์ของผู้ดูจาก Locals (ต้องมี middleware InjectViewer วางก่อนแล้ว)
		v, ok := c.Locals("viewer").(*accessctx.ViewerAccess)
		if !ok || v == nil {
			return fiber.ErrUnauthorized
		}

		items, next, err := repository.ListAllPostsVisibleToViewer(
			c.Context(), client, curStr, limit, v.SubtreeNodeIDs, // ⬅️ ส่งสิทธิ์ของผู้ดู
		)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		resp := dto.ListByCategoryResp[bson.M]{
			Items:      items,
			NextCursor: next,
			HasMore:    next != nil,
		}
		return c.JSON(resp)
	}
}
