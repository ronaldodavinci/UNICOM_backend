package handlers

import (
	"context"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/Software-eng-01204341/Backend/internal/accessctx"
	"github.com/Software-eng-01204341/Backend/model"
)

type FeedRepository interface {
	List(ctx context.Context, opts model.QueryOptions) ([]model.FrontPost, *bson.ObjectID, error)
}

type FeedService struct {
	Repo   FeedRepository
	Client *mongo.Client
}

func NewFeedService(repo FeedRepository, client *mongo.Client) *FeedService {
	return &FeedService{Repo: repo, Client: client}
}

func (s *FeedService) FeedHandler(c *fiber.Ctx) error {
	limit, _ := strconv.ParseInt(c.Query("limit", "20"), 10, 64)
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}

	var until bson.ObjectID
	if cur := c.Query("cursor"); cur != "" {
		if oid, err := bson.ObjectIDFromHex(cur); err == nil {
			until = oid
		}
	}

	// ✅ ใช้ viewer จาก middleware (JWT + InjectViewer)
	// ✅ viewer มาจาก middleware InjectViewer
	// ใช้เพื่อกำหนดสิทธิ์การมองเห็น feed ของ user
	vAny := c.Locals("viewer")
	if vAny == nil {
		return fiber.ErrUnauthorized
	}
	viewer, ok := vAny.(*accessctx.ViewerAccess)
	if !ok || viewer == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "viewer context broken")
	}

	viewerID := viewer.UserID
	allowedNodeIDs := viewer.SubtreeNodeIDs

	opts := model.QueryOptions{
		Roles:          splitCSV(c.Query("role")),     // filter multi-string
		Categories:     splitCSV(c.Query("category")), // filter multi-string
		AuthorIDs:      parseAuthorIDs(splitCSV(c.Query("author"))),
		TextSearch:     c.Query("q"),
		Limit:          limit,
		UntilID:        until,
		ViewerID:       viewerID,
		AllowedNodeIDs: allowedNodeIDs,
	}

	items, next, err := s.Repo.List(c.Context(), opts)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	resp := fiber.Map{"items": items}
	if next != nil {
		resp["next_cursor"] = next.Hex()
	}
	return c.JSON(resp)
}

// ===== helpers =====
func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func parseAuthorIDs(ids []string) []bson.ObjectID {
	if len(ids) == 0 {
		return nil
	}
	out := make([]bson.ObjectID, 0, len(ids))
	for _, h := range ids {
		if oid, err := bson.ObjectIDFromHex(strings.TrimSpace(h)); err == nil {
			out = append(out, oid)
		}
	}
	return out
}
