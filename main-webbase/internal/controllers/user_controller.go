package controllers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"main-webbase/internal/models"
)

// CreateUser godoc
// @Summary Create a new user
// @Description Create user with first name, last name, etc.
// @Tags users
// @Accept json
// @Produce json
// @Param user body models.User true "User Data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /users [post]
func CreateUser(c *fiber.Ctx, client *mongo.Client) error {
	collection := client.Database("big_workspace").Collection("user")

	var user models.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	user.ID = bson.NewObjectID()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, user)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"success": true,
		"message": "User created successfully",
		"data":    user,
	})
}

// GetAllUser godoc
// @Summary Get all users
// @Description Returns all users in database
// @Tags users
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /users [get]
func GetAllUser(c *fiber.Ctx, client *mongo.Client) error {
	collection := client.Database("big_workspace").Collection("user")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// bson.D{} = get all documents
	cursor, err := collection.Find(ctx, bson.D{})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer cursor.Close(ctx)

	// .All = reads all documents and fill into users using cursor
	var users []models.User
	if err := cursor.All(ctx, &users); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Users fetched succesfully",
		"data":    users,
	})
}

// GetUserBy godoc
// @Summary Get user by field
// @Description Get a user by ID, firstname, lastname, etc.
// @Tags users
// @Produce json
// @Param value path string true "Search value"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /user/{field}/{value} [get]
func GetUserBy(c *fiber.Ctx, client *mongo.Client, field string) error {
	collection := client.Database("big_workspace").Collection("user")
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

	var users []models.User
	if err := cursor.All(ctx, &users); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	if len(users) == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "No users found"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    users,
	})
}

// DeleteUser godoc
// @Summary Delete user by ID
// @Description Delete a user with given ID
// @Tags users
// @Param id path string true "User ID"
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /users/{id} [delete]
func DeleteUser(c *fiber.Ctx, client *mongo.Client) error {
	collection := client.Database("big_workspace").Collection("user")
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
		return c.Status(404).JSON(fiber.Map{"error": "User not found"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "User deleted successfully",
	})
}
