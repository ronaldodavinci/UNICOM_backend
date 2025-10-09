package utils

import (
	"time"
	"github.com/gofiber/fiber/v2"
)

const CursorCookieName = "posts_after_rfc3339"

func SetCursorCookie(c *fiber.Ctx, t *time.Time) {
	if t == nil {
		c.Cookie(&fiber.Cookie{
			Name:     CursorCookieName,
			Value:    "",
			HTTPOnly: true,
			Secure:   true,
			SameSite: "Lax",
			MaxAge:   -1,
			Path:     "/",
		})
		return
	}
	c.Cookie(&fiber.Cookie{
		Name:     CursorCookieName,
		Value:    t.UTC().Format(time.RFC3339Nano),
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Lax",
		MaxAge:   24 * 3600,
		Path:     "/",
	})
}

func ReadCursorCookie(c *fiber.Ctx) *time.Time {
	val := c.Cookies(CursorCookieName, "")
	if val == "" {
		return nil
	}
	if tt, err := time.Parse(time.RFC3339Nano, val); err == nil {
		return &tt
	}
	return nil
}

func TimeToRFC3339(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format(time.RFC3339Nano)
}
