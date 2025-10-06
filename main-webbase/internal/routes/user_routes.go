package routes

import (
	"main-webbase/internal/controllers"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutesUser(app *fiber.App) {
	user := app.Group("/users")

	user.Get("/myprofile", controllers.GetMyProfileHandler())
	user.Get("/profile", controllers.GetUserProfileHandler())

	user.Get("/", controllers.GetAllUser())

	// Query by field
	user.Get("/id/:value", controllers.GetUserBy("id"))
	user.Get("/firstname/:value", controllers.GetUserBy("firstname"))
	user.Get("/lastname/:value", controllers.GetUserBy("lastname"))
	user.Get("/thaiprename/:value", controllers.GetUserBy("thaiprename"))
	user.Get("/gender/:value", controllers.GetUserBy("gender"))
	user.Get("/typeperson/:value", controllers.GetUserBy("typeperson"))
	user.Get("/studentid/:value", controllers.GetUserBy("studentid"))
	user.Get("/advisorid/:value", controllers.GetUserBy("advisorid"))

	// Delete user
	app.Delete("/users/:id", controllers.DeleteUser())

}
