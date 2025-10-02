package routes

import (
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"main-webbase/internal/controllers"
)

func SetupRoutesUser_Role(app *fiber.App, client *mongo.Client) {

	// Assign role to user
	app.Post("/users/:user_id/roles", func(c *fiber.Ctx) error {
		return controllers.CreateUserRole(c, client)
	})

	// Remove role assignment
	app.Delete("/users/roles/:id", func(c *fiber.Ctx) error {
		return controllers.DeleteUserRole(c, client)
	})

	// List all roles of a given user
	app.Get("/users/:user_id/roles", func(c *fiber.Ctx) error {
		return controllers.ListOneUserRoles(c, client)
	})

	// List all users with a given role
	app.Get("/roles/:role_id/users", func(c *fiber.Ctx) error {
		return controllers.ListOneRoleUsers(c, client)
	})
}

// example
// curl -X POST http://127.0.0.1:8000/users/63f5e4f4e1b2c4a1d6f8e9b0/roles \
// -H "Content-Type: application/json" \
// -d '{
// 	"role_id": "63f5e4f4e1b2c4a1d6f8e9b0"
// }'

// curl -X DELETE http://127.0.0.1:8000/users/roles/64a5b8c9f1e2d3a4b5c6d7e8

// curl -X GET http://127.0.0.1:8000/users/63f5e4f4e1b2c4a1d6f8e9b0/roles

// curl -X GET http://127.0.0.1:8000/roles/63f5e4f4e1b2c4a1d6f8e9b0/users
