package routes

import (
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"main-webbase/internal/controllers"
)

func SetupRoutesRole(app *fiber.App, client *mongo.Client) {

	app.Post("/roles", func(c *fiber.Ctx) error {
		return controllers.CreateRole(c, client)
	})
	// ตัวอย่าง
	// curl -X POST http://127.0.0.1:8000/roles \
	// -H "Content-Type: application/json" \
	// -d '{
	// "role_name": "admin",
	// "role_path": "/admin",
	// "perm_blog": true,
	// "perm_event": true,
	// "perm_comment": true,
	// "perm_childrole": false,
	// "perm_siblingrole": false
	// }'

	// POSTMAN
	// Method: POST
	// URL: http://127.0.0.1:8000/roles
	// Header: Content-Type: application/json
	// Body: {
	// "role_name": "admin",
	// "role_path": "/admin",
	// "perm_blog": true,
	// "perm_event": true,
	// "perm_comment": true,
	// "perm_childrole": false,
	// "perm_siblingrole": false
	// }

	app.Get("/roles", func(c *fiber.Ctx) error {
		return controllers.GetAllRole(c, client)
	})

	app.Delete("/roles/:id", func(c *fiber.Ctx) error {
		return controllers.DeleteRole(c, client)
	})

	// ตัวอย่าง
	// Query by Attribute
	// เวลาใช้ต้องใส่ curl -X GET http://127.0.0.1:8000/role/role_name/admin

	// POSTMAN
	// Method: GET
	// URL: http://127.0.0.1:8000/role/role_name/admin
	app.Get("/role/id/:value", func(c *fiber.Ctx) error {
		return controllers.GetRoleBy(c, client, "_id")
	})

	app.Get("/role/role_name/:value", func(c *fiber.Ctx) error {
		return controllers.GetRoleBy(c, client, "role_name")
	})

	app.Get("/role/role_path/:value", func(c *fiber.Ctx) error {
		return controllers.GetRoleBy(c, client, "role_path")
	})

	app.Get("/role/perm_blog/:value", func(c *fiber.Ctx) error {
		return controllers.GetRoleBy(c, client, "perm_blog")
	})

	app.Get("/role/perm_event/:value", func(c *fiber.Ctx) error {
		return controllers.GetRoleBy(c, client, "perm_event")
	})

	app.Get("/role/perm_comment/:value", func(c *fiber.Ctx) error {
		return controllers.GetRoleBy(c, client, "perm_comment")
	})

	app.Get("/role/perm_childrole/:value", func(c *fiber.Ctx) error {
		return controllers.GetRoleBy(c, client, "perm_childrole")
	})

	app.Get("/role/PermSiblingRole/:value", func(c *fiber.Ctx) error {
		return controllers.GetRoleBy(c, client, "PermSiblingRole")
	})
}
