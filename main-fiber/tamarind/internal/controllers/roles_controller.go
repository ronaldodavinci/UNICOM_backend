package controllers

import (
	"context"
	"strings"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"tamarind/internal/models"
)

// helper: convert role name into URL-friendly slug
// e.g. "Admin User!" -> "admin-user"
func slugify(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			prevDash = false
			continue
		}
		if r == ' ' || r == '_' || r == '-' {
			if !prevDash {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	out := b.String()
	out = strings.Trim(out, "-")
	if out == "" {
		out = "role"
	}
	return out
}

// CREATE ROLE with auto tree role_path
func CreateRole(c *fiber.Ctx, client *mongo.Client) error {
	collection := client.Database("big_workspace").Collection("role")

	// parse JSON body into Role struct
	var role models.Role
	if err := c.BodyParser(&role); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// generate MongoDB ObjectID
	role.ID = bson.NewObjectID()

	// build role_path = parent/slugified_role_name
	seg := slugify(role.RoleName)
	parent := strings.Trim(role.RolePath, " /")
	if parent != "" {
		role.RolePath = parent + "/" + seg
	} else {
		role.RolePath = seg
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// ensure role_path is unique by checking existing documents
	path := role.RolePath
	n := 1
	for {
		count, err := collection.CountDocuments(ctx, bson.M{"role_path": path})
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		if count == 0 {
			role.RolePath = path
			break
		}
		n++
		path = role.RolePath + "-" + strconv.Itoa(n)
	}

	// insert into MongoDB
	_, err := collection.InsertOne(ctx, role)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// return JSON response
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
		"message": "Roles fetched successfully",
		"data":    roles,
	})
}

// GET ROLE BY FIELD (id, role_name, etc.)
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

	// แก้เพิ่ม
	cursor, err := collection.Find(ctx, filter) // เดิมใช้ FindOne -> Find ดึงข้อมูลได้หลายตัว
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


// GET ONE ROLE BY ID
func GetRoleByID(c *fiber.Ctx, client *mongo.Client) error {
	collection := client.Database("big_workspace").Collection("role")

	// get role_id from URL path
	id := c.Params("id")
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid role ID"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// query MongoDB
	var role models.Role
	err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&role)
	if err == mongo.ErrNoDocuments {
		return c.Status(404).JSON(fiber.Map{"error": "Role not found"})
	}
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// return JSON
	return c.JSON(fiber.Map{
		"role_id":          role.ID.Hex(),
		"role_name":        role.RoleName,
		"role_path":        role.RolePath,
		"perm_blog":        role.PermBlog,
		"perm_event":       role.PermEvent,
		"perm_comment":     role.PermComment,
		"perm_childrole":   role.PermChildRole,
		"perm_siblingrole": role.PermSiblingRole,
	})
}


// DELETE ROLE
func DeleteRole(c *fiber.Ctx, client *mongo.Client) error {
	collection := client.Database("big_workspace").Collection("role")
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
		return c.Status(404).JSON(fiber.Map{"error": "Role not found"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Role deleted successfully",
	})
}

