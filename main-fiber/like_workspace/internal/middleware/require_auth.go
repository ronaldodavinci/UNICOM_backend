package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// RequireAuth checks if request has a user_id in Locals.
// If not -> return 401 Unauthorized.
func RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if v := c.Locals("user_id"); v == nil {
			return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
		} else if uid, ok := v.(string); !ok || strings.TrimSpace(uid) == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
		}
		return c.Next()
	}
}
