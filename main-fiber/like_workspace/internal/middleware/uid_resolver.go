package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// UIDFromLocals ดึง user_id จาก Locals ที่ JWT middleware set ไว้
func UIDFromLocals(c *fiber.Ctx) (string, error) {
	uid, _ := c.Locals("user_id").(string)
	if uid == "" {
		return "", fiber.ErrUnauthorized
	}
	return uid, nil
}
