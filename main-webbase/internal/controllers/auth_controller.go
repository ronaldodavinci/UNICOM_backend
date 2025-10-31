package controllers

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"main-webbase/database"
	"main-webbase/internal/models"
	"net/smtp"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// -------------------- OTP GENERATION --------------------

// GenerateOTP creates a numeric OTP of given length
func GenerateOTP(length int) (string, error) {
	const digits = "0123456789"
	otp := make([]byte, length)
	_, err := rand.Read(otp)
	if err != nil {
		return "", err
	}
	for i := range otp {
		otp[i] = digits[otp[i]%10]
	}
	return string(otp), nil
}

// SendOTPEmail sends an OTP email
func SendOTPEmail(toEmail, otp string) error {
	smtpHost := "smtp.zoho.com"
	smtpPort := "587"
	fromEmail := "otp@kucom.art"
	password := "tcq%ecM1" // ⚠️ move to env var in production

	subject := "Your Registration OTP"
	body := fmt.Sprintf("Hello,\n\nYour OTP code is: %s\nThis code will expire in 5 minutes.\n\n", otp)
	message := fmt.Sprintf("From: %s\nTo: %s\nSubject: %s\n\n%s", fromEmail, toEmail, subject, body)

	auth := smtp.PlainAuth("", fromEmail, password, smtpHost)
	return smtp.SendMail(smtpHost+":"+smtpPort, auth, fromEmail, []string{toEmail}, []byte(message))
}

// -------------------- REGISTER --------------------

// Register godoc
// @Summary Request OTP for registration
// @Description Send OTP to email before creating user
// @Tags auth
// @Accept json
// @Produce json
// @Param registerRequest body models.RegisterRequest true "Register Request"
// @Success 200 {object} map[string]interface{} "OTP sent successfully"
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /register [post]
func Register(c *fiber.Ctx) error {
	collection := database.DB.Collection("users")
	otpCollection := database.DB.Collection("otp_requests")

	var registerRequest models.RegisterRequest
	if err := c.BodyParser(&registerRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if !strings.HasSuffix(registerRequest.Email, "@ku.th") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Email must be a @ku.th address"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check existing user
	var existUser models.User
	err := collection.FindOne(ctx, bson.M{"email": registerRequest.Email}).Decode(&existUser)
	if err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Email already exists"})
	}

	// Generate OTP
	otp, err := GenerateOTP(6)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate OTP"})
	}

	// Send OTP email
	if err := SendOTPEmail(registerRequest.Email, otp); err != nil {
		log.Println("Failed to send OTP email:", err)
	}

	// Store OTP with expiry
	otpCollection.DeleteOne(ctx, bson.M{"email": registerRequest.Email})
	_, err = otpCollection.InsertOne(ctx, bson.M{
		"email":      strings.ToLower(registerRequest.Email),
		"otp":        otp,
		"expires_at": time.Now().Add(10 * time.Minute),
		"user_data":  registerRequest,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to store OTP"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "OTP sent to email.",
	})
}

func VerifyOTP(c *fiber.Ctx) error {
	var req struct {
		Email string `json:"email"`
		OTP   string `json:"otp"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	otpCollection := database.DB.Collection("otp_requests")
	userCollection := database.DB.Collection("users")

	// Delete expired OTPs automatically
	_, _ = otpCollection.DeleteMany(ctx, bson.M{"expires_at": bson.M{"$lt": time.Now()}})

	// Lookup OTP (case-insensitive email)
	var otpRecord bson.M
	if err := otpCollection.FindOne(ctx, bson.M{"email": strings.ToLower(req.Email)}).Decode(&otpRecord); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "No OTP found for this email"})
	}

	// Check OTP match
	storedOTP, ok := otpRecord["otp"].(string)
	if !ok || storedOTP != req.OTP {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid OTP"})
	}

	// ---- EXPIRY CHECK ----
	rawExpiry := otpRecord["expires_at"]
	log.Println("DEBUG: expires_at raw value:", rawExpiry, "type:", fmt.Sprintf("%T", rawExpiry))

	var expiry time.Time
	switch v := rawExpiry.(type) {
	case bson.DateTime:
		expiry = time.UnixMilli(int64(v))
	case bson.RawValue:
		t, ok := v.TimeOK()
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Invalid expiry format"})
		}
		expiry = t
	case time.Time:
		expiry = v
	case string:
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Invalid expiry format"})
		}
		expiry = t
	default:
		log.Println("DEBUG: unknown expiry type:", fmt.Sprintf("%T", v), "value:", v)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Invalid expiry type"})
	}

	if time.Now().After(expiry) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "OTP expired"})
	}
	// ------------------------------------------

	// Convert user_data back to struct safely
	var reg models.RegisterRequest
	bsonBytes, err := bson.Marshal(otpRecord["user_data"])
	if err != nil {
		log.Println("DEBUG: failed to marshal user_data:", err, otpRecord["user_data"])
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "User data corrupted"})
	}
	if err := bson.Unmarshal(bsonBytes, &reg); err != nil {
		log.Println("DEBUG: failed to unmarshal user_data:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "User data corrupted"})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(reg.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to hash password"})
	}

	// Create user
	user := models.User{
		ID:           bson.NewObjectID(),
		FirstName:    reg.FirstName,
		LastName:     reg.LastName,
		ThaiPrefix:   reg.ThaiPrefix,
		Gender:       reg.Gender,
		TypePerson:   reg.TypePerson,
		StudentID:    reg.StudentID,
		AdvisorID:    reg.AdvisorID,
		Email:        reg.Email,
		PasswordHash: string(hashedPassword),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if _, err := userCollection.InsertOne(ctx, user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create user"})
	}

	// Cleanup used OTP
	otpCollection.DeleteOne(ctx, bson.M{"email": strings.ToLower(req.Email)})

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User created successfully.",
		"user_id": user.ID.Hex(),
	})
}

// -------------------- LOGIN --------------------
func Login(c *fiber.Ctx) error {
	var loginRequest models.LoginRequest
	if err := c.BodyParser(&loginRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if !strings.HasSuffix(loginRequest.Email, "@ku.th") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Email must be a @ku.th address"})
	}

	collection := database.DB.Collection("users")
	var user models.User
	filter := bson.M{"email": strings.ToLower(loginRequest.Email)}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := collection.FindOne(ctx, filter).Decode(&user); err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid Email credentials"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database query failed"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(loginRequest.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid Password"})
	}

	claims := jwt.MapClaims{
		"uid": user.ID.Hex(),
		"sub": user.ID.Hex(),
		"exp": time.Now().Add(time.Hour * 72).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
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
