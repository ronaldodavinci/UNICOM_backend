package controllers

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"main-webbase/dto"
	"main-webbase/internal/repository"
	"main-webbase/internal/services"
)

type OrgTreeHandler struct {
	orgRepo *repository.OrgUnitRepository
}

func NewOrgTreeHandler(r *repository.OrgUnitRepository) *OrgTreeHandler {
	return &OrgTreeHandler{orgRepo: r}
}

func CreateOrgUnitHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body dto.OrgUnitDTO
		if err := c.BodyParser(&body); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		node, err := services.CreateOrgUnit(body, ctx)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.Status(fiber.StatusCreated).JSON(dto.OrgUnitReport{
			OrgID:	   	node.ID.Hex(),
			OrgPath: 	node.OrgPath,
			Name:    	node.Name,
			ShortName:  node.ShortName,
		})
	}
}

// GetTree godoc
// @Summary      Get organization unit tree
// @Description  Returns a tree of all organization units, sorted by label. Optional language code for labels.
// @Tags         org
// @Accept       json
// @Produce      json
// @Param        lang query string false "Language code for labels (e.g. 'en', 'ru')"
// @Success      200 {array} dto.OrgUnitNode
// @Failure      500 {object} map[string]interface{}
// @Router       /org/units/tree [get]
func (h *OrgTreeHandler) GetTree(c *fiber.Ctx) error {
	lang := c.Query("lang")

	// simple repo fetch
	orgs, err := h.orgRepo.Find(c.Context(), map[string]any{})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "db error")
	}

	nodes := []*dto.OrgUnitNode{}
	for _, o := range orgs {
		nodes = append(nodes, &dto.OrgUnitNode{
			OrgPath:   o.OrgPath,
			Type:      o.Type,
			Label:     pickLabel(o.Label, lang, o.ShortName, o.OrgPath),
			Labels:    o.Label,
			ShortName: o.ShortName,
			Children:  []*dto.OrgUnitNode{},
		})
	}

	// sort nodes for output
	sort.Slice(nodes, func(i, j int) bool {
		return strings.ToLower(nodes[i].Label) < strings.ToLower(nodes[j].Label)
	})
	return c.JSON(nodes)
}

func pickLabel(labels map[string]string, lang, short, path string) string {
	if v, ok := labels[lang]; ok && v != "" {
		return v
	}
	if short != "" {
		return short
	}
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return path
}
