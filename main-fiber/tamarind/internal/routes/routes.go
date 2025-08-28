// package routes

// import (
// 	"github.com/gofiber/fiber/v2"
// 	"tamarind/internal/handlers"
// 	"go.mongodb.org/mongo-driver/v2/mongo"
// )

// func Register(app *fiber.App, client *mongo.Client) {
// 	app.Get("/hello", handlers.GetHello)

// 	app.Get("/post", func(c *fiber.Ctx) error {
// 		return handlers.GetUserHandler(c, client)
// 	})
// }

package routes

import (
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"tamarind/internal/controllers"
)

func Register(app *fiber.App, client *mongo.Client) {
	// ===== User routes =====
	app.Post("/users", func(c *fiber.Ctx) error { return controllers.CreateUser(c, client) })
	app.Get("/users", func(c *fiber.Ctx) error { return controllers.GetAllUser(c, client) })
	app.Get("/user/:field/:value", func(c *fiber.Ctx) error {
		field := c.Params("field")
		return controllers.GetUserBy(c, client, field)
	})
	app.Delete("/user/:id", func(c *fiber.Ctx) error { return controllers.DeleteUser(c, client) })

	// ===== Role routes =====
	app.Post("/roles", func(c *fiber.Ctx) error { return controllers.CreateRole(c, client) })
	app.Get("/roles", func(c *fiber.Ctx) error { return controllers.GetAllRole(c, client) })
	app.Get("/role/:field/:value", func(c *fiber.Ctx) error {
		field := c.Params("field")
		return controllers.GetRoleBy(c, client, field)
	})
	app.Delete("/role/:id", func(c *fiber.Ctx) error { return controllers.DeleteRole(c, client) })

	// ===== UserRole routes =====
	app.Post("/user_roles", func(c *fiber.Ctx) error { return controllers.CreateUserRole(c, client) })
	app.Get("/user_roles", func(c *fiber.Ctx) error { return controllers.GetAllUserRole(c, client) })
	app.Get("/user_role/:field/:value", func(c *fiber.Ctx) error {
		field := c.Params("field")
		return controllers.GetUserRoleBy(c, client, field)
	})
	app.Delete("/user_role/:id", func(c *fiber.Ctx) error { return controllers.DeleteUserRole(c, client) })
}
