package controllers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"main-webbase/internal/models"
)

// ASSIGN ROLE to USER
func CreateUserRole(c *fiber.Ctx, client *mongo.Client) error {
	collection := client.Database("big_workspace").Collection("user_role")

	userID := c.Params("userID")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing user ID"})
	}

	var body struct {
		RoleID string `json:"role_id"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	userObjID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}
	roleObjID, err := bson.ObjectIDFromHex(body.RoleID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid role ID"})
	}

	userRole := models.User_Role{
		UserID: userObjID,
		RoleID: roleObjID,
	}

	_, err = collection.InsertOne(context.Background(), userRole)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to assign role"})
	}

	return c.Status(fiber.StatusCreated).JSON(userRole)
}

// UNASSIGN ROLE from USER
func DeleteUserRole(c *fiber.Ctx, client *mongo.Client) error {
	collection := client.Database("big_workspace").Collection("user_role")
	id := c.Params("id")

	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	if res.DeletedCount == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "User&Role pair not found"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "User&Role pair deleted successfully",
	})
}
