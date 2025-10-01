package routes

import "github.com/gofiber/fiber/v2"

// SetupRoutes wires all route groups under /api
func SetupRoutes(api fiber.Router) {
	// Auth
	RegisterAuthRoutes(api)

	// Abilities
	RegisterAbilitiesRoutes(api)

	// Events
	RegisterEventRoutes(api)

	// Org Units (tree)
	RegisterOrgRoutes(api)

	// Memberships
	RegisterMembershipRoutes(api)

	// Positions
	RegisterPositionRoutes(api)

	// Policies
	RegisterPolicyRoutes(api)

	// Posts
	RegisterPostRoutes(api)

	// Users
	RegisterUserRoutes(api)
}
