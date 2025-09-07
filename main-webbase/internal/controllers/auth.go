package controllers

import (
	"context"
	"main-webbase/internal/models"
	"os"
	"time"
	// "fmt"
	"golang.org/x/crypto/bcrypt"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Define a struct for login data from the request body
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Login(c *fiber.Ctx, client *mongo.Client) error {
	var loginRequest LoginRequest
	if err := c.BodyParser(&loginRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Find the user in the database by their username
	collection := client.Database("test").Collection("users")
	var user models.User // Assuming you have a User model
	filter := bson.M{"email": loginRequest.Email}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := collection.FindOne(ctx, filter).Decode(&user); err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid Email credentials"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database query failed"})
	}

	// Compare the provided password with the hashed password in the database
	// if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginRequest.Password)); err != nil {
	//  return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid Password"})
	// }

	// if user.PasswordHash != loginRequest.Password {
	// 	fmt.Println("User's stored password:", user.PasswordHash)
	// 	fmt.Println("Provided password:", loginRequest.Password)

	// 	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid Password"})
	// }
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(loginRequest.Password)); err != nil {
		// If the comparison fails, it's a password mismatch
		// fmt.Println("User's stored password:", user.PasswordHash) 
		// fmt.Println("Provided password:", loginRequest.Password)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid Password"})
	}
	// Create JWT Claims
	claims := jwt.MapClaims{
		"Email":   user.Email,
		"user_id": user.ID.Hex(),
		"exp":     time.Now().Add(time.Hour * 72).Unix(), // Token expires in 72 hours
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with a secret key (use an environment variable!)
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "your_strong_secret_key" // Fallback for demonstration, use env var in production
	}
	t, err := token.SignedString([]byte(secret))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not sign token"})
	}

	// Return the user data and the access token
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"user":        user,
		"accessToken": t,
	})
}
