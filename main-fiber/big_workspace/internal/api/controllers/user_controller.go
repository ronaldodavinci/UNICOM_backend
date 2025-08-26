package controllers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"main-fiber/big_workspace/internal/api/models"
)

// POST
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

// GET ALL USERS
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

// GET 1 USERS
func GetUserByID(c *fiber.Ctx, client *mongo.Client) error {
	collection := client.Database("big_workspace").Collection("user")
	id := c.Params("id")

	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "User not found"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    user,
	})
}

// DELETE
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