package routes

import (
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"like_workspace/internal/handlers"
	"like_workspace/internal/repository"
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
	//   curl -X GET http://localhost:3000/api/users
	users := api.Group("/users")

	// Postman:
	//   Method: GET
	//   URL:    http://localhost:3000/api/users
	users.Get("/", func(c *fiber.Ctx) error {
		return handlers.GetUserHandler(c, d.Client)
	})

	// ============================================================
	// Posts
	// ============================================================
	posts := api.Group("/posts")

	// POST /api/posts/blog
	// Example:
	//   curl -X POST http://localhost:3000/api/posts/blog \
	//   -H "Content-Type: application/json" \
	//   -d '{"title":"hello","body":"world"}'
	posts.Post("/blog", handlers.CreatePostHandler(d.Client))
	
	// GET /api/posts/visibility/cursor
	// Example:
	//   curl -X GET http://localhost:3000/api/posts/visibility

	posts.Get("/visibility/cursor", handlers.GetPostsVisibilityCursor(d.Client))
	
	// GET /api/posts/feed
	// Example:
	//   curl -X GET http://localhost:3000/api/posts/feed
	//	 curl -X GET "http://localhost:3000/api/posts/feed?cursor=..."
	//	 curl -X GET "http://localhost:3000/api/posts/feed?tag=..."
	//	 curl -X GET "http://localhost:3000/api/posts/feed?category=..."
	//	 curl -X GET "http://localhost:3000/api/posts/feed?author=..."
	//	 curl -X GET "http://localhost:3000/api/posts/feed?q=..."
	repo := repository.NewMongoFeedRepo(d.Client)
	feedRepo := handlers.NewFeedService(repo)
	posts.Get("/feed", feedRepo.FeedHandler)
	
	// ============================================================
	// Misc
	// ============================================================

	// Health check
	// GET /api/healthz → "ok"
	api.Get("/healthz", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})
}
