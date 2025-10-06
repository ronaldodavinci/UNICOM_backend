package routes

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/Software-eng-01204341/Backend/internal/handlers"
	"github.com/Software-eng-01204341/Backend/internal/repository"
)

// Deps holds shared dependencies to inject into handlers.
type Deps struct {
	Client *mongo.Client
}

// Register mounts all HTTP routes in one place.
// Keep paths lowercase, grouped by resource, and easy to discover.
func Register(app *fiber.App, d Deps) {
	api := app.Group("/api")

	// ============================================================
	// Users
	// ============================================================
	// GET /api/users
	// Example:
	//   curl -X GET http://localhost:8000/api/users
	users := api.Group("/users")

	// Postman:
	//   Method: GET
	//   URL:    http://localhost:8000/api/users
	users.Get("/", func(c *fiber.Ctx) error {
		return handlers.GetUserHandler(c, d.Client)
	})

	// ============================================================
	// Posts
	// ============================================================
	posts := api.Group("/posts")

	// POST /api/posts/blog
	// Example:
	//   curl -X POST http://localhost:8000/api/posts/blog \
	//   -H "Content-Type: application/json" \
	//   -d '{"title":"hello","body":"world"}'
	posts.Post("/blog", handlers.CreatePostHandler(d.Client))

	// GET /api/posts/visibility/cursor
	// Example:
	//   curl -X GET http://localhost:8000/api/posts/visibility/cursor
	posts.Get("/visibility/cursor", handlers.GetPostsVisibilityCursor(d.Client))

	// GET /api/posts/feed
	// ใช้ JWT token + InjectViewer แทนการส่ง ?user=
	// Examples:
	//   curl -X GET "http://localhost:8000/api/posts/feed" -H "Authorization: Bearer <JWT>"
	//   curl -X GET "http://localhost:8000/api/posts/feed?cursor=<hex>" -H "Authorization: Bearer <JWT>"
	//   curl -X GET "http://localhost:8000/api/posts/feed?role=student,/faculty/*" -H "Authorization: Bearer <JWT>"
	//   curl -X GET "http://localhost:8000/api/posts/feed?category=news,events" -H "Authorization: Bearer <JWT>"
	//   curl -X GET "http://localhost:8000/api/posts/feed?author=<id1>,<id2>" -H "Authorization: Bearer <JWT>"
	//   curl -X GET "http://localhost:8000/api/posts/feed?q=keyword" -H "Authorization: Bearer <JWT>"
	repo := repository.NewMongoFeedRepo(d.Client)
	feedSvc := handlers.NewFeedService(repo, d.Client)
	posts.Get("/feed", feedSvc.FeedHandler)

	// ============================================================
	// Debug
	// ============================================================
	// /debug/viewer?user=... ยังไว้ทดสอบได้ (ออปชัน)
	RegisterDebug(app, d.Client)

	// ============================================================
	// Trending
	// ============================================================	
	// GET /api/trending
	// ใช้ JWT token + InjectViewer แทนการส่ง ?user=
	// Examples:
	//   curl -X GET "http://localhost:8000/api/trending/today" -H "Authorization: Bearer <JWT>"
	//   curl -X GET "http://localhost:8000/api/trending/all" -H "Authorization: Bearer <JWT>"
	//   curl -X GET "http://localhost:8000/api/trending/one?tag=..." -H "Authorization: Bearer <JWT>"
	//   curl -X GET "http://localhost:8000/api/trending/posts?tag=..." -H "Authorization: Bearer <JWT>"
	trRepo := repository.NewMongoHashtagTrendingRepoWithDBName(d.Client)
	_ = trRepo.EnsureIndexes(context.Background()) // optional
	trHdl := handlers.NewHashtagTrendingHandler(trRepo)

	trending := api.Group("/trending")
	trending.Get("/today", trHdl.TopToday) // แบบที่ 1
	trending.Get("/all",   trHdl.TopAll)   // แบบที่ 2
	trending.Get("/one",   trHdl.CountOne)
	trending.Get("/posts", trHdl.ListPostsByTag) // แบบที่ 3 (รายละเอียดโพสต์)
	
	// ============================================================
	// Misc
	// ============================================================

	// Health check
	// GET /api/healthz → "ok"
	api.Get("/healthz", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})
}
