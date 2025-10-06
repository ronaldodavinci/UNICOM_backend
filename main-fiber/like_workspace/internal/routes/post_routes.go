package routes

import (
	"github.com/Software-eng-01204341/Backend/internal/handlers"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func PostRoutes(app *fiber.App, client *mongo.Client) {
	post := app.Group("/posts")

	post.Post("/", handlers.CreatePostHandler(client))

	post.Get("/:post_id", handlers.GetIndividualPostHandler(client))

	app.Delete("/posts/:id", handlers.DeletePostHandler(client))
}
