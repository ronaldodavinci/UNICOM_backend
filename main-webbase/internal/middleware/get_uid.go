package middleware

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// UIDFromLocals ดึง user_id จาก Locals ที่ JWT middleware set ไว้
func UIDFromLocals(c *fiber.Ctx) (string, error) {
	uid, _ := c.Locals("user_id").(string)
	if uid == "" {
		return "", fiber.ErrUnauthorized
	}
	fmt.Println("UIDFromLocals: uid=%s\n", uid)
	return uid, nil
}

// UIDObjectID ดึง user_id จาก Locals แล้วแปลงเป็น bson.ObjectID
func UIDObjectID(c *fiber.Ctx) (bson.ObjectID, error) {
	v := c.Locals("user_id")
	uid, ok := v.(string)
	if !ok || uid == "" {
		return bson.NilObjectID, fiber.ErrUnauthorized
	}

	oid, err := bson.ObjectIDFromHex(uid)
	if err != nil {
		return bson.NilObjectID, fiber.ErrUnauthorized
	}
	return oid, nil
}
