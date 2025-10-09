package routes

import (
	"main-webbase/internal/controllers"
	"main-webbase/internal/repository"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// **ยังไม่ได้ทำ JWT/Session + middleware ที่จะรู้ว่า userId = ใคร, role = อะไร จาก token เลยใช้เป็นการ mockผ่านheader
func CommentRoutes(app *fiber.App, client *mongo.Client) {
	db := client.Database("unicom")
	repo := &repository.CommentRepository{
		Client:      client,
		ColComments: db.Collection("comments"),
		ColPosts:    db.Collection("posts"),
	}
	h := &controllers.CommentHandler{Repo: repo}

	// group ของ post → route ที่ขึ้นต้นด้วย /posts/:postId
	posts := app.Group("/posts")

	// GET /posts/:postId/comments
	// ดึงรายการคอมเมนต์ทั้งหมดของโพสต์นั้น
	// รองรับการเรียงลำดับ (ใหม่สุดมาก่อน) + pagination (cursor)
	//
	// Example Request: ขอ 2 comments ที่ถัดจาก cursor xxxx  แต่ถ้าเป็นครั้งแรกส่งมาแค่ limit
	// ครั้งแรก  GET http://localhost:8000/posts/66c62.../comments?limit=2
	// ครั้งต่อไป ใช้cursorที่ได้จากการเรียกครั้งก่อน  GET http://localhost:8000/posts/66c62.../comments?limit=2&cursor=xxxx
	posts.Get("/:postId/comments", h.List)

	// POST /posts/:postId/comments
	// สร้างคอมเมนต์ใหม่ในโพสต์ ระบุด้วย postId
	// Example Request:
	//   POST http://localhost:8000/posts/66c62.../comments
	posts.Post("/:postId/comments", h.Create)

	// PUT /comments/:commentId
	// อัปเดตคอมเมนต์ที่มี id ตรงกับ commentId
	// เจ้าของคอมเมนต์หรือ root สามารถแก้ได้
	// Example Request:
	//   PUT http://localhost:8000/comments/66e4d7b17a12f9dbf8123abc

	app.Put("/comments/:commentId", h.Update)

	// DELETE /comments/:commentId
	// ลบคอมเมนต์ตาม id
	// เฉพาะเจ้าของคอมเมนต์หรือ root เท่านั้นที่สามารถลบได้
	// Example Request:
	//   DELETE http://localhost:8000/comments/66e4d7b17a12f9dbf8123abc
	app.Delete("/comments/:commentId", h.Delete)
}
