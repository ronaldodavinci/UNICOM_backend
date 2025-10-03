package accessctx

import (
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Key สำหรับดึงค่าใน handler
const LocalsKey = "viewer"

// jwtToUserID เป็นฟังก์ชันที่คุณมีอยู่แล้ว (แปลง Bearer → ObjectID)
// ใส่ของจริงของโปรเจกต์คุณแทน
func jwtToUserID(c *fiber.Ctx) (bson.ObjectID, error) {
	// ตัวอย่างคร่าว ๆ
	// token := strings.TrimPrefix(c.Get("Authorization"), "Bearer ")
	// return auth.ParseTokenToUserID(token)
	return bson.ObjectID{}, fiber.ErrUnauthorized
}

func WithViewerAccess(db *mongo.Database) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := jwtToUserID(c)
		if err != nil {
			return fiber.ErrUnauthorized
		}
		va, err := BuildViewerAccess(c.Context(), db, userID)
		if err != nil {
			return fiber.ErrInternalServerError
		}
		c.Locals(LocalsKey, va)
		return c.Next()
	}
}
