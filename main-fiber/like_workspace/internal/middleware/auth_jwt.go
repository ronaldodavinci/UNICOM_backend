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
		UID                  string `json:"uid,omitempty"` // เผื่อบางระบบใส่ uid ตรง ๆ
		jwt.RegisteredClaims        // มี Subject (sub), Exp, Iat, Nbf ฯลฯ
	}

	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" || !strings.HasPrefix(strings.ToLower(auth), "bearer ") {
			return c.Next() // ไม่มี header ก็ปล่อยผ่าน (anonymous)
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
				// อนุญาตเฉพาะ HMAC HS256 เท่านั้น
				if t.Method != jwt.SigningMethodHS256 {
					return nil, fiber.ErrUnauthorized
				}
				return []byte(secret), nil
			},
			jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		)
		if err != nil || !token.Valid {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid token")
		}

		// ดึง uid: ใช้ claims.UID ก่อน ถ้าไม่มีก็ใช้ Subject (sub)
		uid := claims.UID
		if uid == "" {
			uid = claims.Subject
		}
		// สำรองกรณีโทเคนเป็น MapClaims ผสม ๆ มา
		if uid == "" {
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
