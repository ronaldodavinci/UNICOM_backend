package routes

import (
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"main-webbase/internal/controllers"
)

func SetupRoutesUser_Role(app *fiber.App, client *mongo.Client) {

	app.Post("/user_role/:userID", func(c *fiber.Ctx) error {
		return controllers.CreateUserRole(c, client)
	})
	// ตัวอย่าง
	// curl -X POST http://127.0.0.1:8000/user_role/63f5e4f4e1b2c4a1d6f8e9b0 \
	// -H "Content-Type: application/json" \
	// -d '{
	// 	"role_id": "63f5e4f4e1b2c4a1d6f8e9b0"
	// }'

	// POSTMAN
	// Method: POST
	// URL: http://127.0.0.1:8000/user_role/USER_OBJECT_ID
	// Header: Content-Type: application/json
	// Body: raw, JSON
	// {
	// 	"role_id": "ROLE_OBJECT_ID"
	// }

	app.Delete("/user_role/:id", func(c *fiber.Ctx) error {
		return controllers.DeleteUserRole(c, client)
	})
}
