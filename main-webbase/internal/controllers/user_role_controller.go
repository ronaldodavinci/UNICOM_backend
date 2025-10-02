package controllers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"main-webbase/internal/models"
)

// CreateUserRole godoc
// @Summary Assign role to user
// @Description Assign an existing role to a user
// @Tags user_roles
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Param body body models.UserRoleRequest true "Role assignment payload"
// @Success 201 {object} models.User_Role
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /users/{user_id}/roles [post]
func CreateUserRole(c *fiber.Ctx, client *mongo.Client) error {
	collection := client.Database("big_workspace").Collection("user_role")

	userID := c.Params("user_id")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing user ID"})
	}

	// Used for getting input as string (Change to RoleID later)
	var body models.UserRoleRequest

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

// DeleteUserRole godoc
// @Summary Unassign role from user
// @Description Remove a role assignment from a user
// @Tags user_roles
// @Produce json
// @Param id path string true "UserRole ID (document _id)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /users/roles/{id} [delete]
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

// ListOneUserRoles godoc
// @Summary Get all roles for a user
// @Description Retrieve all roles (with details) assigned to a user
// @Tags user_roles
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /users/{user_id}/roles [get]
func ListOneUserRoles(c *fiber.Ctx, client *mongo.Client) error {
	collection := client.Database("big_workspace").Collection("user_role")
	userID := c.Params("user_id")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing user ID"})
	}

	userObjID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Aggregate user_role <-> role collection in mongo
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"user_id": userObjID}}},
		{{
			Key: "$lookup",
			Value: bson.M{
				"from":         "role",
				"localField":   "role_id",
				"foreignField": "_id",
				"as":           "role_details",
			},
		}},
		{{Key: "$unwind", Value: "$role_details"}},
	}
	// Example
	// 1. Find all role in user
	// 2. Join with role collection
	// 	{
	// "_id": "xxxx",
	// "user_id": "123",
	// "role_id": "abc",
	// "role_info": [
	// 	{ "_id": "abc", "name": "admin", "description": "Administrator role" }
	// ]
	// }

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	if len(results) == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "No roles found for this user"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    results,
	})
}

// ListOneRoleUsers godoc
// @Summary Get all users from one role
// @Description Retrieve all users (with details) containing this role
// @Tags user_roles
// @Produce json
// @Param role_id path string true "Role ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /roles/{role_id}/users [get]
func ListOneRoleUsers(c *fiber.Ctx, client *mongo.Client) error {
	collection := client.Database("big_workspace").Collection("user_role")
	roleID := c.Params("role_id")
	if roleID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing role ID"})
	}

	roleObjID, err := bson.ObjectIDFromHex(roleID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid role ID"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Aggregate user_role <-> user collection in mongo
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"role_id": roleObjID}}},
		{{
			Key: "$lookup",
			Value: bson.M{
				"from":         "user",
				"localField":   "user_id",
				"foreignField": "_id",
				"as":           "user_details",
			},
		}},
		{{Key: "$unwind", Value: "$user_details"}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	if len(results) == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "No users found for this role"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    results,
	})
}
