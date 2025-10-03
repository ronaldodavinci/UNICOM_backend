// internal/routes/debug.go
package routes

import (
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"like_workspace/internal/accessctx"
)

func RegisterDebug(app *fiber.App, client *mongo.Client) {
	db := client.Database("lll_workspace")
	app.Get("/debug/viewer", func(c *fiber.Ctx) error {
		user := c.Query("user") // รับ user_id (hex) จาก query
		if user == "" {
			return fiber.NewError(fiber.StatusBadRequest, "missing ?user=<hex ObjectID>")
		}
		uid, err := bson.ObjectIDFromHex(user)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid user ObjectID")
		}

		va, err := accessctx.BuildViewerAccess(c.Context(), db, uid)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		// คืนข้อมูลทั้งหมดให้ดู
		return c.JSON(fiber.Map{
			"userId":          va.UserID.Hex(),
			"memberships":     va.Memberships,
			"subtree_paths":   va.SubtreePaths,
			"subtree_nodeIds": va.SubtreeNodeIDs,
		})
	})
}
