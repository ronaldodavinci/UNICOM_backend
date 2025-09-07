package routes

import (
	"like_workspace/internal/handlers"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func PostRoutes(app *fiber.App, client *mongo.Client) {
	post := app.Group("/posts")

	post.Post("/", handlers.CreatePostHandler(client))
}
