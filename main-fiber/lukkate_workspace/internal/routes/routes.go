package routes

import (
	"github.com/gofiber/fiber/v2"

	"github.com/pllus/main-fiber/lukkate_workspace/internal/handlers"
)

func Register(app *fiber.App) {
	app.Get("/Post", handlers.GetUsers)

	// พวก privateRoutes/middleware ที่คอมเมนต์ไว้อยู่เดิม
	// ถ้าจะใช้ภายหลัง ค่อยย้ายมาใส่ที่นี่
}
