package controllers

import (
	"main-webbase/config"
	"main-webbase/dto"
	"main-webbase/internal/accessctx"
	"main-webbase/internal/middleware"
	"main-webbase/internal/models"
	"main-webbase/internal/repository"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type CommentHandler struct {
	Repo *repository.CommentRepository
}

func viewerFrom(c *fiber.Ctx) *accessctx.ViewerAccess {
	v, _ := c.Locals("viewer").(*accessctx.ViewerAccess)
	return v
}

// เป็น root ไหม: ถ้ามี OrgPath == "/" หรือ (ถ้ามี field is_root ใน User)
// Root/Admin = มี OrgPath == "/"
func isRootByPath(v *accessctx.ViewerAccess) bool {
	if v == nil {
		return false
	}
	for _, m := range v.Memberships {
		if m.OrgPath == "/" {
			return true
		}
	}
	return false
}

// POST /posts/:postId/comments
func (h *CommentHandler) Create(c *fiber.Ctx) error {
	uid, _ := middleware.UIDObjectID(c)

	postID, err := bson.ObjectIDFromHex(c.Params("postId"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid post id"})
	}

	var body dto.CreateCommentReq
	if err := c.BodyParser(&body); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}

	txt := strings.TrimSpace(body.Text)
	if txt == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "text required"})
	}

	com, err := h.Repo.Create(c.Context(), postID, uid, txt)
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

	resp := dto.ListCommentsResp[models.Comment]{
		Comments:   items,
		NextCursor: next,
		HasMore:    next != nil,
	}
	return c.JSON(resp)
}

// PUT /comments/:commentId
func (h *CommentHandler) Update(c *fiber.Ctx) error {
	uid, _ := middleware.UIDObjectID(c)

	cid, err := bson.ObjectIDFromHex(c.Params("commentId"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid comment id"})
	}

	var body dto.CreateCommentReq
	if err := c.BodyParser(&body); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}

	txt := strings.TrimSpace(body.Text)
	if txt == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "text required"})
	}

	isRoot := isRootByPath(viewerFrom(c))

	updated, err := h.Repo.Update(c.Context(), cid, uid, txt, isRoot)
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
	uid, _ := middleware.UIDObjectID(c)

	cid, err := bson.ObjectIDFromHex(c.Params("commentId"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid comment id"})
	}
	isRoot := isRootByPath(viewerFrom(c))
	okDel, err := h.Repo.Delete(c.Context(), cid, uid, isRoot)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if !okDel {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
	}
	return c.SendStatus(http.StatusNoContent)
}
