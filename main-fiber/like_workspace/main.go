package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"

	// "github.com/gofiber/fiber/v2/middleware/cors"
	// "github.com/gofiber/fiber/v2/middleware/logger"
	// "github.com/gofiber/fiber/v2/middleware/recover"
	"like_workspace/database"
	_ "like_workspace/docs"

	// "like_workspace/internal/handlers"
	"like_workspace/bootstrap"
	"like_workspace/internal/middleware"
	"like_workspace/internal/routes"

	"github.com/gofiber/swagger"
	// "like_workspace/internal/handlers"
)

func init() {
	// บังคับให้ค่าจาก .env ทับค่าเดิมใน Environment
	if err := godotenv.Overload(".env"); err != nil {
		log.Println("No .env file found, using system environment variables")
	}
}

func main() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		panic("JWT_SECRET is required")
	}

	// --- MongoDB Connection ---
	client := database.ConnectMongo()
	cfg := database.LoadConfig()
	// defer database.DisconnectMongo()

	if err := bootstrap.EnsureLikeIndexes(client.Database("lll_workspace")); err != nil {
		log.Fatalf("ensure indexes failed: %v", err)
	}

	// --- Fiber App Setup ---
	app := fiber.New()

	// Swagger docs
	app.Get("/docs/*", swagger.HandlerDefault)

	app.Use(func(c *fiber.Ctx) error {
		if uid := c.Get("X-User-ID"); uid != "" {
			c.Locals("user_id", uid)
		}
		if c.Get("X-Is-Root") == "true" {
			c.Locals("is_root", true)
		}
		return c.Next()
	})
	app.Use(middleware.JWTUidOnly())

	// จากนี่ไปต้องมี JWT (หรือถูก mock ด้วย X-User-ID)
	app.Use(middleware.RequireAuth())

	//เอาไว้ตรวจสอบ Locals ว่ามี user_id มั้ย งานจริงใช้ RequireAuth ด้านบนแทน ปิดไว้ก่อน
	// app.Get("/whoami", func(c *fiber.Ctx) error {
	// 	return c.JSON(fiber.Map{
	// 		"user_id": c.Locals("user_id"),
	// 	})
	// })

	// app.Get("/limit", handlers.GetPostsLimit(client))

	// routes.GetUsersHandler(app, client)

	// app.Post("/postblog", handlers.CreatePostHandler(client))

	// // Register routes
	// routes.RegisterRoutes(app, client)

	// app.Get("/posts/limit-role", handlers.GetPostsLimitrole(client))
	// routes.WhoAmIRoutes(app)
	routes.Register(app, routes.Deps{
		Client: client,
	})

	routes.PostRoutes(app, client)

	routes.LikeRoutes(app, client)

	routes.CommentRoutes(app, client)

	log.Printf("listening at http://localhost:%s", cfg.Port)
	if err := app.Listen(":" + os.Getenv("PORT")); err != nil {
		log.Fatal(err)
	}
}
