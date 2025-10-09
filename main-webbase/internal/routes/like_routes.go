package routes

import (
	"main-webbase/internal/controllers"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func LikeRoutes(app *fiber.App, client *mongo.Client) {
	post := app.Group("/likes")

	post.Post("/", controllers.LikeUnlikeHandler(client))

}
