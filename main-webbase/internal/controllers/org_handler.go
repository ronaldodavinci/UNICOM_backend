package controllers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"main-webbase/dto"
	"main-webbase/internal/services"
)

// CreateOrgUnitHandler godoc
// @Summary      Create a new organization unit
// @Description  Creates a new organization unit node in the hierarchy
// @Tags         Org Units
// @Accept       json
// @Produce      json
// @Param        body  body      dto.OrgUnitDTO  true  "Org Unit Data"
// @Success      201   {object}  dto.OrgUnitReport
// @Failure      400   {object}  dto.ErrorResponse "invalid request body"
// @Failure      500   {object}  dto.ErrorResponse "internal server error"
// @Router       /org/units [post]
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

// GetOrgTree godoc
// @Summary      Get organization tree
// @Description  Returns an organization tree starting from a given path and optional depth
// @Tags         Org Units
// @Accept       json
// @Produce      json
// @Param        start  query     string  true   "Starting org path"
// @Param        depth  query     int     false  "Depth of tree to fetch"
// @Success      200    {array}   dto.OrgUnitTree
// @Failure      400   {object}  dto.ErrorResponse "invalid query parameters"
// @Failure      500   {object}  dto.ErrorResponse "internal server error"
// @Router       /org/units/tree [get]
func GetOrgTree() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var query dto.OrgUnitTreeQuery
		if err := c.QueryParser(&query); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid query parameters")
		}

		tree, err := services.BuildOrgTree(context.Background(), query)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(tree)
	}
}