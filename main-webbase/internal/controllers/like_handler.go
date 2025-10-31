package controllers

import (
	"context"
	"main-webbase/dto"
	"main-webbase/internal/services"
	mid "main-webbase/internal/middleware"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// LikeUnlikeHandler godoc
// @Summary      Toggle like on a target
// @Description  Toggle like (หรือ unlike) ให้โพสต์หรือคอมเมนต์ตาม target ที่ส่งมา โดยผูกกับผู้ใช้ที่ล็อกอินอยู่
// @Tags         likes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  dto.LikeRequestDTO  true  "ข้อมูล target ที่จะ like/unlike"
// @Success      200   {object} map[string]interface{}  "สถานะล่าสุดของ like พร้อมจำนวนยอดรวม"
// @Failure      400   {object} dto.ErrorResponse       "ข้อมูลไม่ถูกต้อง"
// @Failure      401   {object} dto.ErrorResponse       "ไม่มีสิทธิ์ (ยังไม่ล็อกอิน)"
// @Failure      500   {object} dto.ErrorResponse       "เกิดข้อผิดพลาดระหว่างประมวลผล"
// @Router       /likes [post]
func LikeUnlikeHandler(client *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user_id, err := mid.UIDObjectID(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).
				JSON(dto.ErrorResponse{Error: "missing userId in context"})
		}

		var body dto.LikeRequestDTO
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Error: "invalid body"})
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		status, payload := services.Like(ctx, client, body, user_id)
		return c.Status(status).JSON(payload)
	}
}
