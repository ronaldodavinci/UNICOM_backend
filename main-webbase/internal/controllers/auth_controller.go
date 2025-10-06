package controllers

import (
	"context"
	"main-webbase/internal/models"
	"main-webbase/database"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Register godoc
// @Summary Register a new user
// @Description Create a new user account with hashed password
// @Tags auth
// @Accept json
// @Produce json
// @Param registerRequest body models.RegisterRequest true "Register Request"
// @Success 201 {object} map[string]interface{} "User registered successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body or email already exists"
// @Failure 500 {object} map[string]interface{} "Failed to create user"
// @Router /register [post]
func Register(c *fiber.Ctx) error {
	collection := database.DB.Collection("users")

	var registerRequest models.RegisterRequest
	if err := c.BodyParser(&registerRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if user exists
	var existUser models.User
	err := collection.FindOne(ctx, bson.M{"email": registerRequest.Email}).Decode(&existUser)
	if err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Email already exists"})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(registerRequest.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to hash password"})
	}

	// Add to db
	user := models.User{
		ID:           bson.NewObjectID(),
		FirstName:    registerRequest.FirstName,
		LastName:     registerRequest.LastName,
		ThaiPrefix:   registerRequest.ThaiPrefix,
		Gender:       registerRequest.Gender,
		TypePerson:   registerRequest.TypePerson,
		StudentID:    registerRequest.StudentID,
		AdvisorID:    registerRequest.AdvisorID,
		Email:        registerRequest.Email,
		PasswordHash: string(hashedPassword),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	_, err = collection.InsertOne(ctx, user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create user"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User registered successfully",
		"user":    user,
	})
}

// Login godoc
// @Summary Login user
// @Description Authenticate user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param loginRequest body models.LoginRequest true "Login Request"
// @Success 200 {object} map[string]interface{} "User and access token"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 401 {object} map[string]interface{} "Invalid credentials"
// @Failure 500 {object} map[string]interface{} "Database or token error"
// @Router /login [post]
func Login(c *fiber.Ctx) error {
	var loginRequest models.LoginRequest
	if err := c.BodyParser(&loginRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Find the user in the database by their username
	collection := database.DB.Collection("users")

	var user models.User
	filter := bson.M{"email": loginRequest.Email}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := collection.FindOne(ctx, filter).Decode(&user); err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid Email credentials"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database query failed"})
	}

	// Compare password and hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(loginRequest.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid Password"})
	}

	// Create JWT Claims
	claims := jwt.MapClaims{
		"uid": user.ID.Hex(), 
		"sub": user.ID.Hex(), // ทำไว้ 2 ชั้น เป็นมาตรฐานเอาไว้ใช้ใน Middleware ด้วย
		"exp": time.Now().Add(time.Hour * 72).Unix(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with a secret key (use an environment variable!)
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Missing JWT_SECRET"})
	}
	
	t, err := token.SignedString([]byte(secret))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not sign token"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"user":        user,
		"accessToken": t,
	})
}
