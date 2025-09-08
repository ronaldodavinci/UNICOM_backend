package handlers

import (
	"net/http"
	"strings"

	"like_workspace/dto"
	"like_workspace/internal/repository"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// GET /api/posts/category/any/cursor?limit=20&cursor=...
func GetPostsInAnyCategoryCursor(client *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		limit := int64(c.QueryInt("limit", 10))
		if limit <= 0 {
			limit = 10
		}
		if limit > 20 {
			limit = 20
		}

		curStr := c.Query("cursor")

		items, next, total, err := repository.ListPostsInAnyCategoryNewestFirst(c.Context(), client, curStr, limit)
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

// GET /api/posts/category/cursor?id=<cid1,cid2,...>&limit=20&cursor=...
func GetPostsByCategoryCursor(client *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// เหมือนคอมเมนต์: limit จาก config
		limit := int64(c.QueryInt("limit", 10))
		if limit <= 0 {
			limit = 10
		}
		if limit > 20 {
			limit = 20
		}
		curStr := c.Query("cursor")

		// ✅ รับได้หลายหมวด (comma-separated) + แปลงด้วย bson.ObjectIDFromHex
		raw := strings.TrimSpace(c.Query("id"))
		if raw == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "missing category id(s) in ?id=",
			})
		}

		parts := strings.Split(raw, ",")
		seen := make(map[bson.ObjectID]struct{}, len(parts))
		cidsB := make([]bson.ObjectID, 0, len(parts))

		for _, s := range parts {
			s = strings.TrimSpace(s)
			if s == "" {
				continue
			}
			oid, err := bson.ObjectIDFromHex(s)
			if err != nil {
				// ตามสไตล์ตัวอย่างของคุณ: ส่ง oid (zero OID เมื่อพัง)
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "invalid category id",
					"id":    oid,
				})
			}
			// ลบซ้ำ
			if _, ok := seen[oid]; !ok {
				seen[oid] = struct{}{}
				cidsB = append(cidsB, oid)
			}
		}
		if len(cidsB) == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "no valid category ids",
			})
		}

		// เรียกเรโพ (ให้เรโพ decode/encode cursor เองตามแพทเทิร์นเพื่อนคุณ)
		items, next, total, err := repository.ListPostsByCategoryNewestFirst(
			c.Context(), client, cidsB, curStr, limit,
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
