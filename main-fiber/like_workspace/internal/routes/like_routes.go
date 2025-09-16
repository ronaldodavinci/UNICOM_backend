package routes

import (
	"like_workspace/internal/handlers"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func LikeRoutes(app *fiber.App, client *mongo.Client) {
	post := app.Group("/likes")

	post.Post("/", handlers.CreateLikeHandler(client))
}