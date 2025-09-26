// middleware/auth.go
package middleware

import (
	"github.com/gofiber/fiber/v2"
	// jwt library ที่คุณใช้
)

func WithUserLocals() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// ตัวอย่างเท่านั้น: สมมติคุณตรวจ JWT แล้วได้ claims
		// จากนั้น set ลง Locals ให้ handler อ่านทีหลัง
		// c.Locals("user_id", claims.UserIDHex)
		// c.Locals("roles", []string{"student","pro"})
		return c.Next()
	}
}
