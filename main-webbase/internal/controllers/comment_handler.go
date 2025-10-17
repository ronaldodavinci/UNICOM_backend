package controllers

import (
	"main-webbase/config"
	"main-webbase/dto"
	"main-webbase/internal/accessctx"
	"main-webbase/internal/middleware"
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

// @Summary      Create a comment
// @Description  Create a new comment under the given post
// @Tags         comments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        postId  path     string                   true  "Post ID (hex ObjectID)"
// @Param        body    body     dto.CreateCommentReq     true  "Comment payload (Text)"
// @Success      201     {object} models.Comment
// @Failure      400     {object} dto.ErrorResponse
// @Failure      401     {object} dto.ErrorResponse
// @Failure      500     {object} dto.ErrorResponse
// @Router       /posts/{postId}/comments [post]
func (h *CommentHandler) Create(c *fiber.Ctx) error {
	uid, _ := middleware.UIDObjectID(c)

	postID, err := bson.ObjectIDFromHex(c.Params("postId"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: "invalid comment id"})
	}

	var body dto.CreateCommentReq
	if err := c.BodyParser(&body); err != nil {
		return c.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: "invalid body"})
	}

	txt := strings.TrimSpace(body.Text)
	if txt == "" {
		return c.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: "text required"})
	}

	com, err := h.Repo.Create(c.Context(), postID, uid, txt)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(dto.ErrorResponse{Error: err.Error()})
	}
	return c.Status(http.StatusCreated).JSON(com)
}

// GET /posts/:postId/comments?limit=20&cursor=...

// @Summary      List comments of a post
// @Description  Fetch comments of a post with cursor pagination
// @Tags         comments
// @Produce      json
// @Param        postId  path   string  true   "Post ID (hex ObjectID)"
// @Param        limit   query  int     false  "Max items per page" minimum(1) maximum(100) default(20)
// @Param        cursor  query  string  false  "Opaque next-page cursor"
// @Success      200     {object} dto.ListCommentsResp
// @Failure      400     {object} dto.ErrorResponse
// @Failure      500     {object} dto.ErrorResponse
// @Router       /posts/{postId}/comments [get]
func (h *CommentHandler) List(c *fiber.Ctx) error {
	postID, err := bson.ObjectIDFromHex(c.Params("postId"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: "invalid post id"})
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
		return c.Status(status).JSON(dto.ErrorResponse{Error: repoErr.Error()})
	}

	viewerID, _ := middleware.UIDObjectID(c)
	anyLiked := false
	if !viewerID.IsZero() {
		// Use likes collection to check if viewer liked any of the returned comments
		likesCol := h.Repo.Client.Database("unicom").Collection("like")
		for _, cm := range items {
			ok, err := repository.CheckIsLiked(c.Context(), likesCol, viewerID, cm.ID, "comment")
			if err == nil && ok {
				anyLiked = true
				break
			}
		}
	}

	resp := dto.ListCommentsResp{
		Comments:   items,
		NextCursor: next,
		HasMore:    next != nil,
		IsLiked:    anyLiked,
	}
	return c.JSON(resp)
}

// PUT /comments/:commentId

// @Summary      Update a comment
// @Description  Only the owner can update a comment
// @Tags         comments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        commentId  path     string                true  "Comment ID (hex ObjectID)"
// @Param        body       body     dto.CreateCommentReq  true  "Fields to update (Text)"
// @Success      200        {object} models.Comment
// @Failure      400        {object} dto.ErrorResponse
// @Failure      401        {object} dto.ErrorResponse
// @Failure      403        {object} dto.ErrorResponse
// @Failure      500        {object} dto.ErrorResponse
// @Router       /comments/{commentId} [put]
func (h *CommentHandler) Update(c *fiber.Ctx) error {
	uid, _ := middleware.UIDObjectID(c)

	cid, err := bson.ObjectIDFromHex(c.Params("commentId"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: "invalid comment id"})
	}

	var body dto.CreateCommentReq
	if err := c.BodyParser(&body); err != nil {
		return c.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: "invalid body"})
	}

	txt := strings.TrimSpace(body.Text)
	if txt == "" {
		return c.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: "text required"})
	}

	updated, err := h.Repo.Update(c.Context(), cid, uid, txt)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(dto.ErrorResponse{Error: err.Error()})
	}
	if updated == nil {
		return c.Status(http.StatusForbidden).JSON(dto.ErrorResponse{Error: "forbidden"})
	}
	return c.JSON(updated)
}

// DELETE /comments/:commentId

// @Summary      Delete a comment
// @Description  Only the owner or an admin (root) can delete a comment
// @Tags         comments
// @Produce      json
// @Security     BearerAuth
// @Param        commentId  path     string  true  "Comment ID (hex ObjectID)"
// @Success      204        {string} string  "no content"
// @Failure      401        {object} dto.ErrorResponse
// @Failure      403        {object} dto.ErrorResponse
// @Failure      404        {object} dto.ErrorResponse
// @Failure      500        {object} dto.ErrorResponse
// @Router       /comments/{commentId} [delete]
func (h *CommentHandler) Delete(c *fiber.Ctx) error {
	uid, _ := middleware.UIDObjectID(c)

	cid, err := bson.ObjectIDFromHex(c.Params("commentId"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: "invalid comment id"})
	}
	isRoot := isRootByPath(viewerFrom(c))
	okDel, err := h.Repo.Delete(c.Context(), cid, uid, isRoot)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(dto.ErrorResponse{Error: err.Error()})
	}
	if !okDel {
		return c.Status(http.StatusForbidden).JSON(dto.ErrorResponse{Error: "forbidden"})
	}
	return c.SendStatus(http.StatusNoContent)
}
