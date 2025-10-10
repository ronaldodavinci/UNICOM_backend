package routes

import (
	"github.com/gofiber/fiber/v2"
	"main-webbase/internal/controllers"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func SetupRoutesEvent(app *fiber.App, client *mongo.Client) {
	event := app.Group("/event")

	// POST /event
	// สร้าง Event ด้วย NodeID ของผู้สร้าง
	// Input จะเป็น
	// 1.รายละเอียดของอีเว้น
	// 		2.List ของวันของ Event []
	event.Post("/", controllers.CreateEventHandler())

	// GET /event
	// ดึงรายการทั้งหมดที่ผู้ใช้ *สามารถ* เห็ยได้โดยดูจาก Organize ของผู้ใช้ทั้งหมดเช็คกับ Status ของ Event
	event.Get("/", controllers.GetAllVisibleEventHandler())

	// DELETE /event/{event_id}
	// ลบ Event โดยดูจาก EventID ที่ส่งเข้ามา
	event.Delete("/:id", controllers.DeleteEventHandler)


	event.Post("/:eventId/qa", controllers.CreateEventQAHandler(client))
}
