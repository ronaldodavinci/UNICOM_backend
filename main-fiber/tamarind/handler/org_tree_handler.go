package handlers

import (
	"sort"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pllus/main-fiber/tamarind/dto"
	"github.com/pllus/main-fiber/tamarind/repositories"
)

type OrgTreeHandler struct {
	orgRepo *repositories.OrgUnitRepository
}

func NewOrgTreeHandler(r *repositories.OrgUnitRepository) *OrgTreeHandler {
	return &OrgTreeHandler{orgRepo: r}
}

// GET /api/org/units/tree
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
