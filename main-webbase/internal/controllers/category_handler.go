package controllers

import (
	"net/http"

	"main-webbase/internal/repository"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// GetCategories ดึงข้อมูลหมวดหมู่ทั้งหมดแบบโง่ๆ (no cursor, no auth, no paging)
func GetCategories(client *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		items, err := repository.ListAllCategories(c.Context(), client)
		if err != nil {
			return c.Status(http.StatusInternalServerError).
				JSON(fiber.Map{"error": err.Error()})
		}
		// ส่งออกเป็น array ธรรมดา:
		// [
		//   {"_id":"66c...7d5","category_name":"Study & Academic","short_name":"study"},
		//   ...
		// ]
		return c.JSON(items)
	}
}
