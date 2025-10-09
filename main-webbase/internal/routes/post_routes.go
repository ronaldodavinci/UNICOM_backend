package routes

import (
	"main-webbase/internal/controllers"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func SetupRoutesPost(app *fiber.App, client *mongo.Client) {
	posts := app.Group("/posts")
	posts.Get("/", controllers.GetPostsVisibilityCursor(client))
	posts.Post("/", controllers.CreatePostHandler(client))
	posts.Get("/:post_id", controllers.GetIndividualPostHandler(client))
	posts.Put("/:post_id", controllers.UpdatePostHandler(client))
	posts.Delete("/:post_id", controllers.DeletePostHandler(client))

}
