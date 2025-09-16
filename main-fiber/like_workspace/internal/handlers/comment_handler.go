package handlers

import (
	"net/http"
	"strings"

	config "like_workspace/configs"
	"like_workspace/dto"
	"like_workspace/internal/repository"
	"like_workspace/model"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type CommentHandler struct {
	Repo *repository.CommentRepository
}

// ดึง user_id จาก middleware (ต้องมีการเซ็ตมาใน c.Locals)
func userIDFrom(c *fiber.Ctx) (bson.ObjectID, bool) {
	if v := c.Locals("user_id"); v != nil {
		if s, ok := v.(string); ok {
			if oid, err := bson.ObjectIDFromHex(s); err == nil {
				return oid, true
			}
		}
	}
	return bson.NilObjectID, false
}

// POST /posts/:postId/comments
func (h *CommentHandler) Create(c *fiber.Ctx) error {
	uid, ok := userIDFrom(c)
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	postID, err := bson.ObjectIDFromHex(c.Params("postId"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid post id"})
	}

	var body dto.CreateCommentReq
	if err := c.BodyParser(&body); err != nil || len(body.Text) == 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "text required"})
	}

	com, err := h.Repo.Create(c.Context(), postID, uid, body.Text)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(http.StatusCreated).JSON(com)
}

// GET /posts/:postId/comments?limit=20&cursor=...
func (h *CommentHandler) List(c *fiber.Ctx) error {
	postID, err := bson.ObjectIDFromHex(c.Params("postId"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid post id"})
	}

	limit := int64(c.QueryInt("limit", config.DefaultLimitComments))
	if limit <= 0 {
		limit = config.DefaultLimitComments
	}
	if limit > config.MaxLimitComments {
		limit = config.MaxLimitComments
	}
	curStr := c.Query("cursor")

	items, next, repoErr := h.Repo.ListByPostNewestFirst(c.Context(), postID, curStr, limit)
	if repoErr != nil {
		status := fiber.StatusInternalServerError
		if strings.Contains(repoErr.Error(), "invalid cursor") {
			status = fiber.StatusBadRequest
		}
		return c.Status(status).JSON(fiber.Map{"error": repoErr.Error()})
	}

	resp := dto.ListCommentsResp[model.Comment]{
		Comments:   items,
		NextCursor: next,
		HasMore:    next != nil,
	}
	return c.JSON(resp)
}

// PUT /comments/:commentId
func (h *CommentHandler) Update(c *fiber.Ctx) error {
	uid, ok := userIDFrom(c)
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	cid, err := bson.ObjectIDFromHex(c.Params("commentId"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid comment id"})
	}

	var body dto.UpdateCommentReq
	if err := c.BodyParser(&body); err != nil || len(body.Text) == 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "text required"})
	}

	isRoot := false
	if v := c.Locals("is_root"); v != nil {
		if b, ok := v.(bool); ok {
			isRoot = b
		}
	}

	updated, err := h.Repo.Update(c.Context(), cid, uid, body.Text, isRoot)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if updated == nil {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
	}
	return c.JSON(updated)
}

// DELETE /comments/:commentId
func (h *CommentHandler) Delete(c *fiber.Ctx) error {
	uid, ok := userIDFrom(c)
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	cid, err := bson.ObjectIDFromHex(c.Params("commentId"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid comment id"})
	}

	okDel, err := h.Repo.Delete(c.Context(), cid, uid, false /* isAdmin */)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if !okDel {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
	}
	return c.SendStatus(http.StatusNoContent)
}
