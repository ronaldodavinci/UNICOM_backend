// internal/routes/debug.go
package routes

import (
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/Software-eng-01204341/Backend/internal/accessctx"
)

func RegisterDebug(app *fiber.App, client *mongo.Client) {
	db := client.Database("lll_workspace")

	// ต้องแน่ใจว่าได้ติดตั้ง middleware.JWTUidOnly() ก่อนถึงเส้นทางนี้
	app.Get("/debug/viewer", func(c *fiber.Ctx) error {
		// ✅ รับ uid/sub จาก JWT ที่ middleware เซ็ตไว้ใน Locals("user_id")
		v := c.Locals("user_id")
		if v == nil {
			return fiber.ErrUnauthorized
		}
		uidStr, ok := v.(string)
		if !ok || uidStr == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid user_id in context")
		}

		// ✅ uid/sub ของคุณเป็น ObjectID (ตามโปรแกรม gen token)
		userOID, err := bson.ObjectIDFromHex(uidStr)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "user_id must be a hex ObjectID")
		}

		// ✅ สร้าง ViewerAccess จาก _id ของผู้ใช้ปัจจุบัน
		va, err := accessctx.BuildViewerAccess(c.Context(), db, userOID)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		// (ถ้าต้องการใช้งานต่อใน chain อื่น ๆ ก็สามารถ c.Locals("viewer", va) ได้)
		// c.Locals("viewer", va)

		// ✅ คืนข้อมูลเพื่อดีบัก
		return c.JSON(fiber.Map{
			"userId":          va.UserID.Hex(),
			"memberships":     va.Memberships,
			"subtree_paths":   va.SubtreePaths,
			"subtree_nodeIds": va.SubtreeNodeIDs,
		})
	})
}
