// internal/handlers/hashtag_trending_handler.go
package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"like_workspace/internal/repository"
)

type HashtagTrendingHandler struct {
	Repo repository.HashtagTrendingRepository
}

func NewHashtagTrendingHandler(repo repository.HashtagTrendingRepository) *HashtagTrendingHandler {
	return &HashtagTrendingHandler{Repo: repo}
}

// GET /api/trending/today?k=10
func (h *HashtagTrendingHandler) TopToday(c *fiber.Ctx) error {
	k, _ := strconv.Atoi(c.Query("k", "10"))
	out, err := h.Repo.TopPublicHashtagsToday(c.Context(), k)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(out)
}

// GET /api/trending/all?k=10
func (h *HashtagTrendingHandler) TopAll(c *fiber.Ctx) error {
	k, _ := strconv.Atoi(c.Query("k", "10"))
	out, err := h.Repo.TopPublicHashtagsAllTime(c.Context(), k)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(out)
}

// GET /api/trending/one?tag=%23golang[&day=YYYY-MM-DD]
func (h *HashtagTrendingHandler) CountOne(c *fiber.Ctx) error {
	tag := c.Query("tag")
	if tag == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "tag is required"})
	}
	day := c.Query("day") // optional
	out, err := h.Repo.CountPublicPostsByHashtag(c.Context(), tag, day)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(out)
}

func (h *HashtagTrendingHandler) ListPostsByTag(c *fiber.Ctx) error {
	tag := c.Query("tag")
	if tag == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "tag is required"})
	}
	day := c.Query("day")

	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	cursor := c.Query("cursor", "")

	items, next, err := h.Repo.ListPublicPostsByHashtag(c.Context(), tag, day, limit, cursor)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	resp := fiber.Map{
		"items": items,
	}
	if next != nil {
		hex := next.Hex()
		resp["nextCursor"] = hex
		resp["next"] = fiber.Map{"cursor": hex} // เผื่อ front เอาไปต่อ url ง่าย ๆ
	}

	return c.JSON(resp)
}
