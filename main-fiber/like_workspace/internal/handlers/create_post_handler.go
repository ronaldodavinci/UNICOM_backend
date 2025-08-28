package handlers

import (
	"time"

	"like_workspace/dto"
	"like_workspace/model"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func CreatePostHandler(client *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// แปลง body เป็น DTO
		var body dto.CreatePostDTO
		if err := c.BodyParser(&body); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid body")
		}

		// ตรวจ validation พื้นฐาน
		if body.UserID == "" || body.RoleID == "" || body.PostText == "" {
			return fiber.NewError(fiber.StatusBadRequest, "userId, roleId and postText are required")
		}
		if len(body.CategoryIDs) == 0 {
			return fiber.NewError(fiber.StatusBadRequest, "at least one categoryId is required")
		}

		// แปลง id จาก string → ObjectID
		userID, err := bson.ObjectIDFromHex(body.UserID)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid userId")
		}
		roleID, err := bson.ObjectIDFromHex(body.RoleID)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid roleId")
		}

		now := time.Now().UTC()

		// สร้าง post object
		post := model.Post{
			UserID:    userID,
			RoleID:    roleID,
			PostText:  body.PostText,
			Picture:   body.PictureUrl, // *string ถ้า nil จะเก็บเป็น null ใน MongoDB
			Video:     body.VideoUrl,
			CreatedAt: now,
			UpdatedAt: now,
		}

		// เลือก DB และ Collection (ไม่ใช้ global)
		db := client.Database("lll_workspace")
		postsCol := db.Collection("posts")
		postCatsCol := db.Collection("post_categories")

		// Insert post
		res, err := postsCol.InsertOne(c.Context(), post)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		post.ID = res.InsertedID.(bson.ObjectID)

		// เตรียม post_categories สำหรับแนบหมวด
		docs := make([]interface{}, 0, len(body.CategoryIDs))
		for i, cidStr := range body.CategoryIDs {
			cid, err := bson.ObjectIDFromHex(cidStr)
			if err != nil {
				// ❌ ถ้า category ผิด → rollback
				_, _ = postsCol.DeleteOne(c.Context(), bson.M{"_id": post.ID})
				return fiber.NewError(fiber.StatusBadRequest, "invalid categoryId: "+cidStr)
			}
			docs = append(docs, model.PostCategory{
				PostID:     post.ID,
				CategoryID: cid,
				OrderIndex: i + 1,
			})
		}

		// 📌 Insert category relations (unordered เพื่อไม่หยุดถ้า duplicate)
		if _, err := postCatsCol.InsertMany(
			c.Context(),
			docs,
			options.InsertMany().SetOrdered(false),
		); err != nil {
			_, _ = postsCol.DeleteOne(c.Context(), bson.M{"_id": post.ID})
			return fiber.NewError(fiber.StatusInternalServerError, "failed to attach categories: "+err.Error())
		}

		return c.Status(fiber.StatusCreated).JSON(post)
	}
}
