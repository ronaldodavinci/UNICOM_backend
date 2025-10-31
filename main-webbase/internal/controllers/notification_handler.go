package controllers

import (
	"context"
	"net/http"
	"time"

	"main-webbase/database"
	"main-webbase/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// GetUnreadNotifications godoc
// @Summary      List unread notifications for the current user
// @Description  Return all unread notifications and the total count for the authenticated user.
// @Tags         notifications
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object} map[string]interface{}  "Unread notification count and list"
// @Failure      500  {object} map[string]string       "Failed to fetch notifications"
// @Router       /notifications [get]
// GET /notifications
func GetUnreadNotifications() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// ดึง userID จาก middleware ที่ inject ไว้หลัง login แล้ว
		userIDVal, _ := middleware.UIDObjectID(c)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		col := database.DB.Collection("notification")
		cursor, err := col.Find(ctx, bson.M{
			"user_id": userIDVal,
			"read":    false,
		})
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to fetch notifications",
			})
		}
		defer cursor.Close(ctx)

		var notifications []bson.M
		if err := cursor.All(ctx, &notifications); err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to parse notifications",
			})
		}

		return c.Status(http.StatusOK).JSON(fiber.Map{
			"unread_count": len(notifications),
			"data":         notifications,
		})
	}
}

// GetNotificationAndMarkRead godoc
// @Summary      Get a notification and mark it as read
// @Description  Fetch a notification by ID for the authenticated user, mark it as read, and return the updated document.
// @Tags         notifications
// @Produce      json
// @Security     BearerAuth
// @Param        id   path   string  true  "Notification ID (hex ObjectID)"
// @Success      200  {object} map[string]interface{}  "Updated notification document"
// @Failure      404  {object} map[string]string       "Notification not found"
// @Failure      500  {object} map[string]string       "Failed to update notification"
// @Router       /notifications/{id} [get]
// GET /notifications/:id
// ใช้ตอนผู้ใช้ "กดดูรายละเอียด" แจ้งเตือน -> mark read แล้วคืนรายละเอียดฉบับเต็ม
func GetNotificationAndMarkRead() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, _ := middleware.UIDObjectID(c)

		hex := c.Params("id")
		notiID, _ := bson.ObjectIDFromHex(hex)

		col := database.DB.Collection("notification")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		filter := bson.M{"_id": notiID, "user_id": userID}
		update := bson.M{"$set": bson.M{
			"read":    true,
			"read_at": time.Now().UTC(),
		}}

		// คืนเอกสารหลังอัปเดต (อ่านง่ายสำหรับหน้า "รายละเอียด")
		opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

		var notif bson.M
		err := col.FindOneAndUpdate(ctx, filter, update, opts).Decode(&notif)
		if err == mongo.ErrNoDocuments {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "notification not found"})
		}
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update notification"})
		}

		return c.Status(http.StatusOK).JSON(fiber.Map{
			"data": notif,
		})
	}
}
