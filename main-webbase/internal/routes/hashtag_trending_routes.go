package routes

import (
	"main-webbase/internal/controllers"
	"main-webbase/internal/repository"
	"context"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// GET /api/trending
	// ใช้ JWT token + InjectViewer แทนการส่ง ?user=
	// Examples:
	//   curl -X GET "http://localhost:8000/api/trending/today" -H "Authorization: Bearer <JWT>"
	//   curl -X GET "http://localhost:8000/api/trending/all" -H "Authorization: Bearer <JWT>"
	//   curl -X GET "http://localhost:8000/api/trending/one?tag=..." -H "Authorization: Bearer <JWT>"
	//   curl -X GET "http://localhost:8000/api/trending/posts?tag=..." -H "Authorization: Bearer <JWT>"
	
func SetupRoutesTrending(app *fiber.App, client *mongo.Client) {
	trRepo := repository.NewMongoHashtagTrendingRepoWithDBName(client)
	_ = trRepo.EnsureIndexes(context.Background()) // optional
	trHdl := controllers.NewHashtagTrendingHandler(trRepo)

	trending := app.Group("/trending")
	trending.Get("/today", trHdl.TopToday) // แบบที่ 1
	trending.Get("/all",   trHdl.TopAll)   // แบบที่ 2
	trending.Get("/one",   trHdl.CountOne)
	trending.Get("/posts", trHdl.ListPostsByTag) // แบบที่ 3 (รายละเอียดโพสต์)
}