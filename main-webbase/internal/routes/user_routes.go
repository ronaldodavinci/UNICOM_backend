package routes

import (
	"main-webbase/internal/controllers"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func SetupRoutesUser(app *fiber.App, client *mongo.Client) {

	app.Get("/users", func(c *fiber.Ctx) error {
		return controllers.GetAllUser(c, client)
	})
	// ตัวอย่าง
	// curl -X GET http://127.0.0.1:8000/users

	app.Delete("/users/:id", func(c *fiber.Ctx) error {
		return controllers.DeleteUser(c, client)
	})
	// ตัวอย่าง
	// curl -X DELETE http://127.0.0.1:8000/users/USER_OBJECT_ID

	// Query by Attribute
	// เวลาใช้ต้องใส่ /user/id/63f5e4f4e1b2c4a1d6f8e9b0
	app.Get("/user/id/:value", func(c *fiber.Ctx) error {
		return controllers.GetUserBy(c, client, "_id")
	})

	app.Get("/user/firstname/:value", func(c *fiber.Ctx) error {
		return controllers.GetUserBy(c, client, "firstname")
	})

	app.Get("/user/lastname/:value", func(c *fiber.Ctx) error {
		return controllers.GetUserBy(c, client, "lastname")
	})

	app.Get("/user/thaiprename/:value", func(c *fiber.Ctx) error {
		return controllers.GetUserBy(c, client, "thaiprename")
	})

	app.Get("/user/gender/:value", func(c *fiber.Ctx) error {
		return controllers.GetUserBy(c, client, "gender")
	})

	app.Get("/user/typeperson/:value", func(c *fiber.Ctx) error {
		return controllers.GetUserBy(c, client, "typeperson")
	})

	app.Get("/user/studentid/:value", func(c *fiber.Ctx) error {
		return controllers.GetUserBy(c, client, "studentid")
	})

	app.Get("/user/advisorid/:value", func(c *fiber.Ctx) error {
		return controllers.GetUserBy(c, client, "advisorid")
	})
	app.Post("/auth/login", func(c *fiber.Ctx) error {
		return controllers.Login(c, client)
	})
}
