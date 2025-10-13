package controllers

import (
    "context"
    "time"

    "github.com/gofiber/fiber/v2"
    "go.mongodb.org/mongo-driver/v2/bson"
    "go.mongodb.org/mongo-driver/v2/mongo/options"
    "main-webbase/dto"
    "main-webbase/internal/services"
    "main-webbase/database"
    "strconv"
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

// ListOrgUnits godoc
// @Summary      List organization units
// @Description  Returns a flat list of org units, optionally filtered by start prefix and search text
// @Tags         Org Units
// @Accept       json
// @Produce      json
// @Param        start   query string false "Org prefix"
// @Param        search  query string false "Search by name or path (case-insensitive)"
// @Param        limit   query int    false "Limit results (default 50)"
// @Success      200    {array}   dto.OrgUnitTree
// @Router       /org/units [get]
func ListOrgUnits() fiber.Handler {
    return func(c *fiber.Ctx) error {
        start := c.Query("start")
        search := c.Query("search")
        limit := int64(50)
        if v := c.Query("limit"); v != "" {
            // ignore parse error -> keep default
            if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 500 { limit = int64(n) }
        }

        filter := bson.M{"status": bson.M{"$ne": "inactive"}}
        if start != "" {
            filter["$or"] = []bson.M{
                {"ancestors": start},
                {"org_path": start},
            }
        }
        if search != "" {
            rx := bson.M{"$regex": search, "$options": "i"}
            and := []bson.M{filter}
            and = append(and, bson.M{"$or": []bson.M{{"name": rx}, {"shortname": rx}, {"org_path": rx}}})
            filter = bson.M{"$and": and}
        }

        cur, err := database.DB.Collection("org_units").Find(c.Context(), filter, options.Find().SetLimit(limit))
        if err != nil { return fiber.NewError(fiber.StatusInternalServerError, err.Error()) }
        defer cur.Close(c.Context())

        type Row struct {
            OrgPath   string `bson:"org_path" json:"org_path"`
            Name      string `bson:"name" json:"name"`
            ShortName string `bson:"shortname" json:"shortname"`
            Type      string `bson:"type" json:"type"`
        }
        out := make([]Row, 0, 50)
        for cur.Next(c.Context()) {
            var r Row
            if err := cur.Decode(&r); err == nil { out = append(out, r) }
        }
        return c.JSON(out)
    }
}
