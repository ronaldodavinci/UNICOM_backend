package routes

import (
	"github.com/gofiber/fiber/v2"
	"main-webbase/internal/controllers"
	"main-webbase/internal/middleware"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func SetupRoutesPost(app *fiber.App, client *mongo.Client) {
	db := client.Database("unicom") 
	posts := app.Group("/posts")
	posts.Use(middleware.InjectViewer(db))
	posts.Get("/", controllers.GetPostsVisibilityCursor(client))
}
