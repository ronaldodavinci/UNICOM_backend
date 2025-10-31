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

// ใช้แทน dto.ListByCategoryResp[bson.M] สำหรับ Swagger (ไม่ generic)
type FeedCursorEnvelope struct {
	Items      []map[string]any `json:"items"`
	NextCursor *string          `json:"next_cursor" example:"eyJpZCI6IjY4ZTYzMDAyZGYyMjVjZDk1NTE3M2RiIn0="`
	HasMore    bool             `json:"has_more" example:"true"`
}

// @Summary      Get posts visible to the viewer
// @Description  List posts that the viewer has permission to see, based on their organizational unit and position.
// @Tags         feed
// @Produce      json
// @Security     BearerAuth
// @Param        limit   query  int     false  "Max items per page" minimum(1) maximum(20) default(10)
// @Param        cursor  query  string  false  "Opaque next-page cursor (base64)"
// @Success      200     {object} controllers.FeedCursorEnvelope
// @Failure      401     {object} dto.ErrorResponse
// @Failure      500     {object} dto.ErrorResponse
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
			return c.Status(http.StatusInternalServerError).JSON(dto.ErrorResponse{Error: err.Error()})
		}

		resp := dto.ListByCategoryResp[bson.M]{
			Items:      items,
			NextCursor: next,
			HasMore:    next != nil,
		}
		return c.JSON(resp)
	}
}
