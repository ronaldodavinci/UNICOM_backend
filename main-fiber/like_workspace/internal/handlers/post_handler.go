package handlers

import (
	"context"
	"errors"
	"time"

	"like_workspace/dto"
	"like_workspace/services"
	mid "like_workspace/internal/middleware"
	repo "like_workspace/internal/repository"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// POST /posts

// CreatePostHandler godoc
// @Summary      Create a post
// @Description  Create a new post with categories, visibility and media URLs
// @Tags         posts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header    string             true  "Bearer {token}"
// @Param        X-Request-Id   header    string             false "Idempotency key"
// @Param        data           body      dto.CreatePostDTO  true  "Post payload"
// @Success      201            {object}  dto.PostResponse
// @Failure      400            {object}  dto.ErrorResponse
// @Failure      401            {object}  dto.ErrorResponse
// @Failure      404            {object}  dto.ErrorResponse
// @Failure      500            {object}  dto.ErrorResponse
// @Router       /posts [post]
func CreatePostHandler(client *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, _ := mid.UIDFromLocals(c)


		var body dto.CreatePostDTO
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).
				JSON(dto.ErrorResponse{Message: "invalid body"})
		}

		// --- basic validation ---
		if body.PostText == "" {
			return c.Status(fiber.StatusBadRequest).
				JSON(dto.ErrorResponse{Message: "postText is required"})
		}
		if body.PostAs.OrgPath == "" || body.PostAs.PositionKey == "" {
			return c.Status(fiber.StatusBadRequest).
				JSON(dto.ErrorResponse{Message: "postAs.org_path and postAs.position_key are required"})
		}

		// default visibility
		if body.Visibility.Access == "" {
			body.Visibility.Access = "public"
		}

		// Use a proper context.Context for Mongo (ไม่ใช้ c.Context() เพราะไม่ใช่ Go context)
		ctx := context.Background()

		post, err := services.CreatePostWithMeta(client, userID, body, ctx)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrUserNotFound):
				return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: "user not found"})
			case errors.Is(err, services.ErrOrgNodeNotFound):
				return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "org_path not found"})
			case errors.Is(err, services.ErrPositionNotFound):
				return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "position_key not found"})
			default:
				// อาจเช็ค duplicate key, validation อื่น ๆ เพิ่มได้หาก service ตีกลับ error เฉพาะ
				return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: err.Error()})
			}
		}

		return c.Status(fiber.StatusCreated).JSON(post)
	}
}

// GET /posts/:post_id

// GetIndividualPostHandler godoc
// @Summary      Get a post detail
// @Description  Return post detail (user, position, org path, visibility, categories, likes count, etc.)
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        post_id  path  string  true  "Post ID (hex)"
// @Success      200  {object}  dto.PostResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /posts/{post_id} [get]
func GetIndividualPostHandler(client *mongo.Client) fiber.Handler {
	const dbName = "lll_workspace"

	return func(c *fiber.Ctx) error {
		postIDHex := c.Params("post_id")
		if postIDHex == "" {
			return fiber.NewError(fiber.StatusBadRequest, "missing post_id in route")
		}
		postID, err := bson.ObjectIDFromHex(postIDHex)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid post_id")
		}

		// context พร้อม timeout
		ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
		defer cancel()

		db := client.Database(dbName)

		resp, err := services.GetPostDetail(ctx, db, postID)
		if err != nil {
			// ถ้าถูก wrap ด้วย %w จาก service จะเช็ค ErrNoDocuments ได้
			if errors.Is(err, mongo.ErrNoDocuments) {
				return fiber.NewError(fiber.StatusNotFound, "post not found")
			}
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		return c.Status(fiber.StatusOK).JSON(resp)
	}
}

// DeletePostHandlerWithClient สร้าง Fiber handler โดยรับ mongo.Client + dbName
// @Summary      Soft delete post (status: active -> inactive)
// @Description  เปลี่ยนสถานะโพสต์จาก active เป็น inactive (soft delete)
// @Tags         posts
// @Param        id   path      string  true  "Post ID (hex)"
// @Success      204
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /posts/{id} [delete]
func DeletePostHandler(client *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		idParam := c.Params("id")
		postID, err := bson.ObjectIDFromHex(idParam)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid post id")
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		db := client.Database("lll_workspace")

		err = repo.DeletePost(db, postID, ctx)
		if err != nil {
			if err.Error() == "post not found or already inactive" {
				return fiber.NewError(fiber.StatusNotFound, err.Error())
			}
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		return c.SendStatus(fiber.StatusNoContent)
	}
}
