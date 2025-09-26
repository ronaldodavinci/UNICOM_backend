package handlers

import (
	"context"
	"errors"

	"like_workspace/dto"
	"like_workspace/services"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// CreatePostHandler godoc
// @Summary      Create a post
// @Description  Create a new post with categories, visibility, and posting role.
// @Description  Requirements:
// @Description   • Auth via Bearer token
// @Description   • body.postText required
// @Description   • body.postAs.org_path & body.postAs.position_key required
// @Tags         posts
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string              true  "Bearer <JWT>"
// @Param        data           body      dto.CreatePostDTO   true  "Post payload"
// @Success      201            {object}  dto.PostResponse    "Created"
// @Failure      400            {object}  dto.ErrorResponse   "Bad request (missing fields / invalid org_path or position_key)"
// @Failure      401            {object}  dto.ErrorResponse   "Unauthorized (missing/invalid token)"
// @Failure      404            {object}  dto.ErrorResponse   "User not found"
// @Failure      500            {object}  dto.ErrorResponse   "Internal server error"
// @Router       /posts [post]
func CreatePostHandler(client *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := services.FetchUserID(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).
				JSON(dto.ErrorResponse{Message: "missing userId in context"})
		}

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
