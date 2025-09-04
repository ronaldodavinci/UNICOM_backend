package handlers

import (
	"fmt"
	"regexp"
	"time"

	"like_workspace/dto"
	"like_workspace/model"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func ExtractHashtags(text string) []string {
	re := regexp.MustCompile(`#([\p{L}\p{M}0-9_]+)`)

	result := re.FindAllString(text, -1)

	return result
}

// CreatePostHandler godoc
// @Summary      Create a post
// @Description  Create a new blog post and attach categories
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        body  body      CreatePostRequest  true  "post payload"
// @Success      201   {object}  Post
// @Failure      400   {string}  string
// @Failure      500   {string}  string
// @Router       /PostBlog [post]

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
			CreatedAt: now,
			UpdatedAt: now,
		}

		// เลือก DB และ Collection
		db := client.Database("lll_workspace")
		postsCol := db.Collection("posts")

		// Insert post
		res, err := postsCol.InsertOne(c.Context(), post)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		post.ID = res.InsertedID.(bson.ObjectID)

		// Insert Hashtags
		hashtagsCol := db.Collection("hashtags")
		hashtags := ExtractHashtags(body.PostText)

		fmt.Println("Extracted hashtags:", hashtags) // debug log

		if len(hashtags) > 0 {
			hashtagDocs := make([]interface{}, 0, len(hashtags))
			dateOnly := post.CreatedAt.Format("2006-01-02")
			for _, tag := range hashtags {
				hashtagDocs = append(hashtagDocs, model.PostHashtag{
					PostID: post.ID,
					Tag:    tag,
					Date:   dateOnly,
				})
			}
			if _, err := hashtagsCol.InsertMany(
				c.Context(),
				hashtagDocs,
				options.InsertMany().SetOrdered(false),
			); err != nil {
				fmt.Println("failed to insert hashtags:", err)
			}
		}

		if len(body.CategoryIDs) == 0 {
			return c.Status(fiber.StatusCreated).JSON(post)
		}
		if len(body.RoleIDs) == 0 {
			return c.Status(fiber.StatusCreated).JSON(post)
		}

		postCatsCol := db.Collection("post_categories")
		postRoleVisCol := db.Collection("post_role_visibility")

		// เตรียม post_categories สำหรับแนบหมวด
		// category_docs เป็น slice เพื่อรอไป insert ลง category
		category_docs := make([]interface{}, 0, len(body.CategoryIDs)) // สร้าง slice -> make(data_type, length_default, lenth_max)
		for i, cidStr := range body.CategoryIDs {
			cid, err := bson.ObjectIDFromHex(cidStr)
			if err != nil {
				// ถ้า category ผิด → rollback
				_, _ = postsCol.DeleteOne(c.Context(), bson.M{"_id": post.ID})
				return fiber.NewError(fiber.StatusBadRequest, "invalid roleId: "+cidStr)
			}
			category_docs = append(category_docs, model.PostCategory{
				PostID:     post.ID,
				CategoryID: cid,
				OrderIndex: i + 1,
			})
		}

		// Insert category relations (unordered เพื่อไม่หยุดถ้า duplicate)
		if _, err := postCatsCol.InsertMany(
			c.Context(),
			category_docs,
			options.InsertMany().SetOrdered(false),
		); err != nil {
			_, _ = postsCol.DeleteOne(c.Context(), bson.M{"_id": post.ID})
			return fiber.NewError(fiber.StatusInternalServerError, "failed to attach categories: "+err.Error())
		}

		// Insert role visibility
		role_visible_docs := make([]interface{}, 0, len(body.RoleIDs))
		for _, ridStr := range body.RoleIDs {
			rid, err := bson.ObjectIDFromHex(ridStr)
			if err != nil {
				// ถ้า role ผิด → rollback
				_, _ = postRoleVisCol.DeleteOne(c.Context(), bson.M{"_id": post.ID})
				_, _ = postCatsCol.DeleteMany(c.Context(), bson.M{"_id": post.ID})
				return fiber.NewError(fiber.StatusBadRequest, "invalid roleId: "+ridStr)
			}
			role_visible_docs = append(role_visible_docs, model.PostRoleVisibility{
				PostID: post.ID,
				RoleID: rid,
			})
		}

		if _, err := postRoleVisCol.InsertMany(
			c.Context(),
			role_visible_docs,
			options.InsertMany().SetOrdered(false),
		); err != nil {
			_, _ = postRoleVisCol.DeleteOne(c.Context(), bson.M{"_id": post.ID})
			return fiber.NewError(fiber.StatusInternalServerError, "failed to attach roles: "+err.Error())
		}

		return c.Status(fiber.StatusCreated).JSON(post)
	}
}
