// // handlers/feed_handler.go (หรือไฟล์ที่กำกับ endpoint /feed)
// package handlers

// import (
// 	"context"
// 	"strconv"

// 	"github.com/gofiber/fiber/v2"
// 	"go.mongodb.org/mongo-driver/v2/bson"

// 	"like_workspace/model"
// )

// type FeedRepository interface {
// 	List(ctx context.Context, opts model.QueryOptions) ([]model.FrontPost, *bson.ObjectID, error)
// }

// type FeedService struct {
// 	Repo FeedRepository
// }

// func NewFeedService(repo FeedRepository) *FeedService { return &FeedService{Repo: repo} }

// func (s *FeedService) FeedHandler(c *fiber.Ctx) error {
// 	limit, _ := strconv.ParseInt(c.Query("limit", "20"), 10, 64)
// 	var until bson.ObjectID
// 	if cur := c.Query("cursor"); cur != "" {
// 		until, _ = bson.ObjectIDFromHex(cur)
// 	}

// 	// TODO: ดึง roles ผู้ชมจาก context/middleware ของคุณ
// 	var viewerRoles []string
// 	if v := c.Locals("roles"); v != nil {
// 		if rr, ok := v.([]string); ok {
// 			viewerRoles = rr
// 		}
// 	}
	
// 	opts := model.QueryOptions{
// 		TextSearch: c.Query("q"),
// 		Tags:       parseCSV(c.Query("tag")),
// 		Categories: parseCSV(c.Query("category")),
// 		Limit:      limit,
// 		UntilID:    until,
// 		Roles:      viewerRoles,
// 	}

// 	items, next, err := s.Repo.List(c.Context(), opts)
// 	if err != nil {
// 		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
// 	}

// 	resp := fiber.Map{"items": items}
// 	if next != nil {
// 		resp["next_cursor"] = next.Hex()
// 	}
// 	return c.JSON(resp)
// }

// func parseCSV(q string) []string {
// 	if q == "" { return nil }
// 	res := []string{}
// 	start := 0
// 	for i := 0; i <= len(q); i++ {
// 		if i == len(q) || q[i] == ',' {
// 			part := q[start:i]
// 			for len(part) > 0 && part[0] == ' ' { part = part[1:] }
// 			for len(part) > 0 && part[len(part)-1] == ' ' { part = part[:len(part)-1] }
// 			if part != "" { res = append(res, part) }
// 			start = i+1
// 		}
// 	}
// 	return res
// }

// handlers/feed_handler.go
package handlers

import (
	"context"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"

	"like_workspace/model"
)

type FeedRepository interface {
	List(ctx context.Context, opts model.QueryOptions) ([]model.FrontPost, *bson.ObjectID, error)
}

type FeedService struct {
	Repo FeedRepository
}

func NewFeedService(repo FeedRepository) *FeedService { return &FeedService{Repo: repo} }

func (s *FeedService) FeedHandler(c *fiber.Ctx) error {
	limit, _ := strconv.ParseInt(c.Query("limit", "20"), 10, 64)

	// cursor/จนถึงไอดีนี้ (เลื่อนลง)
	var until bson.ObjectID
	if cur := c.Query("cursor"); cur != "" {
		until, _ = bson.ObjectIDFromHex(cur)
	}

	// ===== ViewerID จาก query (?viewer_id=...) =====
	var viewerID bson.ObjectID
	if vid := c.Query("viewer_id"); vid != "" {
		if oid, err := bson.ObjectIDFromHex(vid); err == nil {
			viewerID = oid
		}
	}

	opts := model.QueryOptions{
		TextSearch: c.Query("q"),
		Tags:       splitCSV(c.Query("tag")),
		Categories: splitCSV(c.Query("category")),
		Limit:      limit,
		UntilID:    until,
		ViewerID:   viewerID, // 👈 ใช้ไอดีผู้ใช้แทน roles
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
