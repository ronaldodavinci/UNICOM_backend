package controllers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"tamarind/internal/models"
)

// ASSIGN ROLE TO USER
func CreateUserRole(c *fiber.Ctx, client *mongo.Client) error {
	collection := client.Database("big_workspace").Collection("user_role")

	var userRole models.User_Role
	if err := c.BodyParser(&userRole); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	userRole.ID = bson.NewObjectID()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, userRole)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"success": true,
		"message": "UserRole created successfully",
		"data":    userRole,
	})
}

// GET ALL USER_ROLES
func GetAllUserRole(c *fiber.Ctx, client *mongo.Client) error {
	collection := client.Database("big_workspace").Collection("user_role")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.D{})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer cursor.Close(ctx)

	var userRoles []models.User_Role
	if err := cursor.All(ctx, &userRoles); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "UserRoles fetched successfully",
		"data":    userRoles,
	})
}

// GET USER_ROLE BY FIELD (id, user_id, role_id)
func GetUserRoleBy(c *fiber.Ctx, client *mongo.Client, field string) error {
	collection := client.Database("big_workspace").Collection("user_role")
	value := c.Params("value")

	var filter bson.M
	if field == "_id" {
		objID, err := bson.ObjectIDFromHex(value)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
		}
		filter = bson.M{"_id": objID}
	} else {
		filter = bson.M{field: value}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var userRole models.User_Role
	err := collection.FindOne(ctx, filter).Decode(&userRole)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "UserRole not found"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    userRole,
	})
}

// DELETE USER_ROLE
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
		return c.Status(404).JSON(fiber.Map{"error": "UserRole not found"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "UserRole deleted successfully",
	})
}
