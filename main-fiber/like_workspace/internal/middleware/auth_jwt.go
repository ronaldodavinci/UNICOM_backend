// internal/middleware/auth_jwt.go
package middleware

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func JWTUidOnly() fiber.Handler {
	secret := os.Getenv("JWT_SECRET")

	type MyClaims struct {
		UID string `json:"uid,omitempty"`
		jwt.RegisteredClaims
	}

	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" || !strings.HasPrefix(strings.ToLower(auth), "bearer ") {
			return c.Next()
		}
		if secret == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing JWT_SECRET")
		}

		tokenStr := strings.TrimSpace(auth[7:])
		var claims MyClaims

		token, err := jwt.ParseWithClaims(
			tokenStr,
			&claims,
			func(t *jwt.Token) (any, error) {
				// ✅ เช็คด้วย Alg() แทน pointer
				if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
					return nil, fiber.NewError(fiber.StatusUnauthorized, "unsupported alg")
				}
				return []byte(secret), nil
			},
			jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		)
		if err != nil || !token.Valid {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid token")
		}

		// ✅ รองรับทั้ง uid และ sub
		uid := claims.UID
		if uid == "" {
			uid = claims.Subject
		}
		if uid == "" {
			// เผื่อกรณีเคลมเป็น MapClaims
			if mc, ok := token.Claims.(jwt.MapClaims); ok {
				if v, ok := mc["uid"].(string); ok && v != "" {
					uid = v
				} else if v, ok := mc["sub"].(string); ok && v != "" {
					uid = v
				}
			}
		}
		if uid == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing uid/sub")
		}

		c.Locals("user_id", uid)
		return c.Next()
	}
}
