package handlers

import (
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"like_workspace/internal/repository"
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
func GetPostsLimitrole(client *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := userIDFromCtx(c)
		roles := rolesFromCtx(c)

		results, err := repository.FetchPostsVisible(c.Context(), client, userID, roles, 5)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(results)
	}
}
