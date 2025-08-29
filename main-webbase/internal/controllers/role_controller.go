package controllers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"main-webbase/internal/models"
)

// CREATE ROLE
func CreateRole(c *fiber.Ctx, client *mongo.Client) error {
	collection := client.Database("big_workspace").Collection("role")

	var role models.Role
	if err := c.BodyParser(&role); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	role.ID = bson.NewObjectID()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, role)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"success": true,
		"message": "Role created successfully",
		"data":    role,
	})
}

// GET ALL ROLES
func GetAllRole(c *fiber.Ctx, client *mongo.Client) error {
	collection := client.Database("big_workspace").Collection("role")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.D{})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer cursor.Close(ctx)

	var roles []models.Role
	if err := cursor.All(ctx, &roles); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Roles feteched successfully",
		"data":    roles,
	})
}

// GET ROLE BY PATH
// func GetRoleByPath(c *fiber.Ctx, client *mongo.Client, rolePath string) error {
	
// }


// GET ROLE BY FIELD



// DELETE ROLE BY ID
func DeleteRole(c *fiber.Ctx, client *mongo.Client) error {
	collection := client.Database("big_workspace").Collection("role")
	id := c.Params("id")

	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Invalid ID"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	if res.DeletedCount == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Role not found"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Role deleted successfully",
	})
}