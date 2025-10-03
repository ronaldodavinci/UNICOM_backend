// internal/handlers/feed_handler.go
package handlers

import (
	"context"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"like_workspace/model"
	"like_workspace/internal/accessctx"
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

	var until bson.ObjectID
	if cur := c.Query("cursor"); cur != "" {
		if oid, err := bson.ObjectIDFromHex(cur); err == nil {
			until = oid
		}
	}

	// viewerID จาก query หรือจาก auth locals
	var viewerID bson.ObjectID
	if qs := c.Query("user"); qs != "" {
		if oid, err := bson.ObjectIDFromHex(strings.TrimSpace(qs)); err == nil {
			viewerID = oid
		}
	}
	if viewerID == (bson.ObjectID{}) {
		if uidHex, _ := c.Locals("userId").(string); uidHex != "" {
			if oid, err := bson.ObjectIDFromHex(strings.TrimSpace(uidHex)); err == nil {
				viewerID = oid
			}
		}
	}

	// ดึง subtree node_ids ของผู้ชม (สำหรับ visibility)
	var allowedNodeIDs []bson.ObjectID
	if viewerID != (bson.ObjectID{}) {
		if va, err := accessctx.BuildViewerAccess(c.Context(), s.Client.Database("lll_workspace"), viewerID); err == nil && va != nil {
			allowedNodeIDs = va.SubtreeNodeIDs
		}
	}

	opts := model.QueryOptions{
		Roles:          splitCSV(c.Query("role")),         // 👈 กรองแบบสตริงหลายค่า
		Categories:     splitCSV(c.Query("category")),     // 👈 กรองแบบสตริงหลายค่า
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
