package routes

// import (
// 	"github.com/gofiber/fiber/v2"
// 	"main-webbase/internal/controllers"
// 	"main-webbase/internal/repository"
// 	"main-webbase/internal/services"
// )

// func SetupRoutesAbility(app *fiber.App) {
// 	mRepo := repository.NewMembershipRepository()
// 	pRepo := repository.NewPolicyRepository()

// 	authzService := services.NewAuthzService(mRepo, pRepo)
// 	h := controllers.NewAbilitiesHandler(authzService)

// 	abilities := app.Group("/abilities")
// 	abilities.Get("/:userId", h.GetAbilities)
// }
