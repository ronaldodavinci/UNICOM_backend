package handlers

import (
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// ===== Helpers: ดึงค่าจาก Context =====
func rolesFromCtx(c *fiber.Ctx) []string {
	if v := c.Locals("roles"); v != nil {
		if rs, ok := v.([]string); ok {
			return rs
		}
		if r, ok := v.(string); ok && r != "" {
			return []string{r}
		}
	}
	return []string{}
}
func userIDFromCtx(c *fiber.Ctx) string {
	if v := c.Locals("userID"); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// ===== Handler เดิม: ปรับให้กรองด้วย post_rolevisibility =====
func GetPostsLimitrole(client *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		const limit int64 = 5
		userID := userIDFromCtx(c)
		roles  := rolesFromCtx(c)

		posts := client.Database("lll_workspace").Collection("posts")

		// Pipeline:
		// 1) join rolevisibility -> rv
		// 2) ดึง array roles ตัวแรกของเอกสาร visibility (กรณี 1:1 ต่อโพสต์)
		// 3) match ด้วยกติกา:
		//    - rv_roles ว่าง (ทุกคนเห็น)
		//    - setIntersection(rv_roles, roles ของผู้ใช้) มีสมาชิก > 0
		//    - เป็นเจ้าของโพสต์ (author_id == userID)
		// 4) sort + limit
		pipeline := mongo.Pipeline{
			{{Key: "$lookup", Value: bson.M{
				"from":         "post_rolevisibility",
				"localField":   "_id",
				"foreignField": "post_id",
				"as":           "rv",
			}}},
			// กรณีมีหลายแถวใน rv (ปกติควรมีแถวเดียว) เลือกอันแรก
			{{Key: "$addFields", Value: bson.M{
				"rv_roles": bson.M{"$ifNull": bson.A{
					bson.M{"$let": bson.M{
						"vars": bson.M{"firstRv": bson.M{"$first": "$rv"}},
						"in":   bson.M{"$ifNull": bson.A{"$$firstRv.roles", bson.A{}}},
					}},
					bson.A{},
				}},
			}}},
			// ใช้ $expr เพื่อเขียนเงื่อนไขกับ $size/$setIntersection
			{{Key: "$match", Value: bson.M{
				"$expr": bson.M{
					"$or": bson.A{
						bson.M{"$eq": bson.A{bson.M{"$size": "$rv_roles"}, 0}},
						bson.M{"$gt": bson.A{bson.M{"$size": bson.M{
							"$setIntersection": bson.A{"$rv_roles", roles},
						}}, 0}},
						bson.M{"$eq": bson.A{"$author_id", userID}},
					},
				},
			}}},
			{{Key: "$sort", Value: bson.D{{Key: "timestamp", Value: -1}}}},
			{{Key: "$limit", Value: limit}},
		}

		cursor, err := posts.Aggregate(c.Context(), pipeline, options.Aggregate())
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
