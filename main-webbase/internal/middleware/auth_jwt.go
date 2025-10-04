package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type MyClaims struct {
	UID string `json:"uid,omitempty"`
	jwt.RegisteredClaims
}

func JWTUidOnly(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" || !strings.HasPrefix(strings.ToLower(auth), "bearer ") {
			return c.Next()
		}
		
		tokenStr := strings.TrimSpace(auth[7:])
		var claims MyClaims

		token, err := jwt.ParseWithClaims(
			tokenStr,
			&claims,
			func(t *jwt.Token) (any, error) {
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

		uid := claims.UID
		if uid == "" {
			uid = claims.Subject
		}
		if uid == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing uid")
		}

		c.Locals("user_id", uid)
		return c.Next()
	}
}