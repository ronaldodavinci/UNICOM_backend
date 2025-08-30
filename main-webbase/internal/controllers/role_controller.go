package controllers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"main-webbase/internal/models"
)

// CreateRole godoc
// @Summary Create a new role
// @Description Create a new role with permissions
// @Tags roles
// @Accept json
// @Produce json
// @Param role body models.Role true "Role object"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /roles [post]
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

// GetAllRole godoc
// @Summary Get all roles
// @Description Retrieve all roles in the system
// @Tags roles
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /roles [get]
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

// GetRoleBy godoc
// @Summary Get role by field
// @Description Get a role by ID, name, or path
// @Tags roles
// @Produce json
// @Param field path string true "Field to search (e.g., _id, name, path)"
// @Param value path string true "Value to search for"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /roles/{field}/{value} [get]
func GetRoleBy(c *fiber.Ctx, client *mongo.Client, field string) error {
	collection := client.Database("big_workspace").Collection("role")
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

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer cursor.Close(ctx)

	var roles []models.Role
	if err := cursor.All(ctx, &roles); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	if len(roles) == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "No roles found"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    roles,
	})
}

// DeleteRole godoc
// @Summary Delete role by ID
// @Description Delete a role from the system by its ID
// @Tags roles
// @Produce json
// @Param id path string true "Role ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /roles/{id} [delete]
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
