package controllers

import (
	//"context"
	//"time"

	"main-webbase/internal/repository"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// GetUserHandler godoc
// @Summary Get all users
// @Description Get all users from MongoDB User collection
// @Tags users
// @Accept  json
// @Produce  json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /Post [get]

func GetUserHandler(c *fiber.Ctx, client *mongo.Client) error {
	users, err := repository.FetchUsers(c.Context(), client)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
			"data":    nil,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Users fetched successfully",
		"data":    users,
	})
}
