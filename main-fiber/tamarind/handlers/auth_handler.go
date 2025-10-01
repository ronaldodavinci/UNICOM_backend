package handlers

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pllus/main-fiber/tamarind/dto"
	"github.com/pllus/main-fiber/tamarind/repositories"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	userRepo *repositories.UserRepository
	jwtKey   []byte
}

func NewAuthHandler(u *repositories.UserRepository, key []byte) *AuthHandler {
	return &AuthHandler{userRepo: u, jwtKey: key}
}

// POST /api/auth/login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" || req.Password == "" {
		return fiber.NewError(fiber.StatusBadRequest, "email and password required")
	}

	u, err := h.userRepo.FindByEmail(c.Context(), email)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid credentials")
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)) != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid credentials")
	}

	claims := jwt.MapClaims{
		"sub":   u.ID.Hex(),
		"email": u.Email,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString(h.jwtKey)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "sign failed")
	}

	return c.JSON(dto.LoginResponse{AccessToken: signed})
}

// GET /api/auth/me
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	auth := c.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return fiber.NewError(fiber.StatusUnauthorized, "missing token")
	}
	tokenStr := strings.TrimPrefix(auth, "Bearer ")
	claims := jwt.MapClaims{}
	t, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return h.jwtKey, nil
	})
	if err != nil || !t.Valid {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid token")
	}
	return c.JSON(claims)
}
