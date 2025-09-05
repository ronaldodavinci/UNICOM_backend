package handlers

import "github.com/gofiber/fiber/v2"

// WhoAmI shows the current logged-in user and roles
func WhoAmI() fiber.Handler {
	return func(c *fiber.Ctx) error {
		uid := c.Locals("userID")
		roles := c.Locals("roles")
		return c.JSON(fiber.Map{
			"userID": uid,
			"roles":  roles,
		})
	}
}
