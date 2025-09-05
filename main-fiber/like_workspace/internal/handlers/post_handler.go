package handlers

import (
	//"log"

	"like_workspace/internal/repository"
	"like_workspace/internal/utils"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// GetPostsLimit godoc
// @Summary Get latest posts
// @Description Fetch latest posts with limit=5
// @Tags posts
// @Produce json
// @Success 200 {array} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/posts/limit [get]
func GetPostsLimit(client *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		const limit int64 = 5 // 🔒 locked limit

		collection := client.Database("lll_workspace").Collection("posts")
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

// GetPostsLimitrole godoc
// @Summary Get posts visible to user
// @Description Posts filtered by role visibility or ownership
// @Tags posts
// @Produce json
// @Success 200 {array} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/posts/limit-role [get]
// func GetPostsLimitrole(client *mongo.Client) fiber.Handler {
// 	return func(c *fiber.Ctx) error {
// 		userID := userIDFromCtx(c)
// 		roles := rolesFromCtx(c)

// 		results, err := repository.FetchPostsVisible(c.Context(), client, userID, roles, 5)
// 		if err != nil {
// 			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
// 		}
// 		return c.JSON(results)
// 	}
// }

// func GetFirstPostRaw(client *mongo.Client) fiber.Handler {
// 	return func(c *fiber.Ctx) error {
// 		var doc bson.M
// 		if err := client.Database("lll_workspace").
// 			Collection("posts").
// 			FindOne(c.Context(), bson.M{}).
// 			Decode(&doc); err != nil {
// 			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
// 		}
// 		return c.JSON(doc) // ดูว่า _id เป็นแบบไหน
// 	}
// }


// func GetPostsAllPlainRawHandler(client *mongo.Client) fiber.Handler {
// 	return func(c *fiber.Ctx) error {
// 		items, err := repository.FetchPostsAllPlainRaw(c.Context(), client, 10)
// 		if err != nil { return c.Status(500).JSON(fiber.Map{"error": err.Error()}) }
// 		return c.JSON(fiber.Map{"items": items})
// 	}
// }

const lockedLimit int64 = 5

func GetPostsLimitrole(client *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := userIDFromCtx(c)
		roles := rolesFromCtx(c)

		items, nextCursor, err := repository.FetchPostsVisibleCursor(c.Context(), client, userID, roles, lockedLimit, nil)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		// เซ็ตคุ้กกี้ cursor ถ้ามีหน้าถัดไป
		utils.SetCursorCookie(c, nextCursor)

		return c.JSON(fiber.Map{
			"items":       items,
			"next_cursor": utils.TimeToRFC3339(nextCursor), // client จะไม่ต้องใช้ก็ได้ แต่อยากโชว์ไว้
		})
	}
}

// หน้าถัดไป: ก็ไม่รับพารามิเตอร์ อ่าน cursor จากคุ้กกี้อย่างเดียว
func GetPostsLimitroleNext(client *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := userIDFromCtx(c)
		roles := rolesFromCtx(c)

		after := utils.ReadCursorCookie(c)

		items, nextCursor, err := repository.FetchPostsVisibleCursor(c.Context(), client, userID, roles, lockedLimit, after)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		// อัปเดตคุ้กกี้ (หรือเคลียร์ถ้าไม่มีหน้าถัดไป)
		utils.SetCursorCookie(c, nextCursor)

		return c.JSON(fiber.Map{
			"items":       items,
			"next_cursor": utils.TimeToRFC3339(nextCursor),
		})
	}
}