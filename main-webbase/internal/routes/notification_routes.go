package routes

import (
	"main-webbase/internal/controllers"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func NotificationRoutes(app *fiber.App, client *mongo.Client) {
	noti := app.Group("/notifications")
	noti.Get("/", controllers.GetUnreadNotifications())
	noti.Get("/:id", controllers.GetNotificationAndMarkRead())
}
