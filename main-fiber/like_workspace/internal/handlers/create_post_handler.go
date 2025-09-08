package handlers

import (
	"like_workspace/dto"

	"like_workspace/services"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// CreatePostHandler godoc
// @Summary      Create a post
// @Description  Create a new blog post and attach categories
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        data  body      dto.CreatePostDTO  true  "post payload"
// @Success      201   {object}  dto.PostResponse
// @Failure      400   {object}  dto.ErrorResponse
// @Failure      500   {object}  dto.ErrorResponse
// @Router       /posts [post]
func CreatePostHandler(client *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body dto.CreatePostDTO

		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
				Message: "invalid body",
			})
		}

		if body.UserID == "" || body.RoleID == "" || body.PostText == "" {
			return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
				Message: "userId, roleId and postText are required",
			})
		}

		post, err := services.CreatePostWithMeta(client, body, c.Context())
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
				Message: err.Error(),
			})
		}

		return c.Status(fiber.StatusCreated).JSON(post)
	}
}
