package routes

import (
	"like_workspace/internal/handlers"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func PostRoutes(app *fiber.App, client *mongo.Client) {
	post := app.Group("/posts")

	post.Post("/", handlers.CreatePostHandler(client))

	// ดึงโพสต์เดี่ยวตาม post_id
	post.Get("/:post_id", handlers.GetIndividualPostHandler(client))
}
