package routes

import (
    "github.com/gofiber/fiber/v2"
    "go.mongodb.org/mongo-driver/v2/mongo"

    "like_workspace/internal/handlers"
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
	users := api.Group("/users")

	// GET /api/users
	// Example:
	//   curl -X GET http://localhost:3000/api/users
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
	// GET /api/posts/limit
	// Example:
	//   curl -X GET http://localhost:3000/api/posts/limit
	posts.Get("/limit", handlers.GetPostsLimit(d.Client))

	// GET /api/posts/limit-role
	// Example:
	//   curl -X GET http://localhost:3000/api/posts/limit-role
	posts.Get("/limit-role", handlers.GetPostsLimitrole(d.Client))
	posts.Get("/limit-role/next", handlers.GetPostsLimitroleNext(d.Client))
	
	// GET /api/posts/category/cursor
	// Example:
	//   curl -X GET http://localhost:3000/api/posts/category
	posts.Get("/category/cursor", handlers.GetPostsInAnyCategoryCursor(d.Client))
	posts.Get("/category/wanted/cursor", handlers.GetPostsByCategoryCursor(d.Client))


	// WhoAmI debug
	// GET /api/whoami
	// Example:
	//   curl -X GET http://localhost:3000/api/whoami
	api.Get("/whoami", handlers.WhoAmI())
	
	// ============================================================
	// Misc
	// ============================================================

	// Health check
	// GET /api/healthz → "ok"
	api.Get("/healthz", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})
}
