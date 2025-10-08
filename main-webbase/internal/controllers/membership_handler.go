package controllers

import (
    "github.com/gofiber/fiber/v2"
    "main-webbase/internal/models"
    repo "main-webbase/internal/repository"
    "main-webbase/database"
    "go.mongodb.org/mongo-driver/v2/bson"
    "context"
)

// CreateMembership godoc
// @Summary      Create a new membership
// @Description  Assigns a user to an organization and position
// @Tags         Memberships
// @Accept       json
// @Produce      json
// @Param        body  body      models.MembershipRequestDTO  true  "Membership data"
// @Success      200   {object}  models.MembershipRequestDTO "membership created"
// @Failure      400   {object}  dto.ErrorResponse "invalid body"
// @Failure      500   {object}  dto.ErrorResponse "internal server error"
// @Router       /memberships [post]
func CreateMembership() fiber.Handler {
	return func (c *fiber.Ctx) error {
		var req models.MembershipRequestDTO
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid body")
		}
		if err := repo.InsertMembership(c.Context(), req); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(fiber.Map{"message": "membership created", "data": req})
	}
}

// ListMemberships 
// @Summary      List memberships
// @Description  Returns a list of memberships
// @Tags         memberships
// @Accept       json
// @Produce      json
// @Success      200 {object} map[string][]string
// @Failure      500 {object} map[string]interface{}
// @Router       /memberships [get]
// func (h *MembershipHandler) ListMemberships(c *fiber.Ctx) error {
// 	mems, err := h.membershipRepo.FindAll(c.Context())
// 	if err != nil {
// 		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
// 	}
// 	return c.JSON(mems)
// }

// ListMembershipsWithUsers
// @Summary      List memberships at an org with user details
// @Description  Returns active memberships at exact org_path, joined with basic user info
// @Tags         memberships
// @Accept       json
// @Produce      json
// @Param        org_path query string true "Organization path"
// @Param        active   query string false "active|all (default: active)"
// @Success      200 {array} map[string]interface{}
// @Failure      400 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /memberships/users [get]
func ListMembershipsWithUsers() fiber.Handler {
    return func(c *fiber.Ctx) error {
        orgPath := c.Query("org_path")
        if orgPath == "" {
            return fiber.NewError(fiber.StatusBadRequest, "org_path is required")
        }
        active := c.Query("active", "active")

        colMem := database.DB.Collection("memberships")
        filter := bson.M{"org_path": orgPath}
        if active != "all" {
            filter["active"] = true
        }

        ctx := context.Background()
        cur, err := colMem.Find(ctx, filter)
        if err != nil {
            return fiber.NewError(fiber.StatusInternalServerError, err.Error())
        }
        defer cur.Close(ctx)

        type Row struct {
            ID          bson.ObjectID `bson:"_id" json:"_id"`
            UserID      bson.ObjectID `bson:"user_id" json:"user_id"`
            OrgPath     string        `bson:"org_path" json:"org_path"`
            PositionKey string        `bson:"position_key" json:"position_key"`
            Active      bool          `bson:"active" json:"active"`
        }

        var mems []Row
        if err := cur.All(ctx, &mems); err != nil {
            return fiber.NewError(fiber.StatusInternalServerError, err.Error())
        }
        if len(mems) == 0 {
            return c.JSON([]any{})
        }

        // collect user ids
        ids := make([]bson.ObjectID, 0, len(mems))
        idSet := map[string]struct{}{}
        for _, m := range mems {
            k := m.UserID.Hex()
            if _, ok := idSet[k]; !ok {
                idSet[k] = struct{}{}
                ids = append(ids, m.UserID)
            }
        }

        colUsers := database.DB.Collection("users")
        ucur, err := colUsers.Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
        if err != nil {
            return fiber.NewError(fiber.StatusInternalServerError, err.Error())
        }
        defer ucur.Close(ctx)

        type User struct {
            ID         bson.ObjectID `bson:"_id" json:"_id"`
            FirstName  string        `bson:"firstname" json:"firstname"`
            LastName   string        `bson:"lastname" json:"lastname"`
            Email      string        `bson:"email" json:"email"`
            StudentID  string        `bson:"student_id" json:"student_id"`
        }
        users := map[string]User{}
        for ucur.Next(ctx) {
            var u User
            if err := ucur.Decode(&u); err == nil {
                users[u.ID.Hex()] = u
            }
        }

        // build output
        out := make([]fiber.Map, 0, len(mems))
        for _, m := range mems {
            u := users[m.UserID.Hex()]
            out = append(out, fiber.Map{
                "_id":          m.ID.Hex(),
                "org_path":     m.OrgPath,
                "position_key": m.PositionKey,
                "active":       m.Active,
                "user_id":      m.UserID.Hex(),
                "user": fiber.Map{
                    "_id":        u.ID.Hex(),
                    "firstname":  u.FirstName,
                    "lastname":   u.LastName,
                    "email":      u.Email,
                    "student_id": u.StudentID,
                },
            })
        }
        return c.JSON(out)
    }
}

// DeactivateMembership godoc
// @Summary      Deactivate membership
// @Description  Set membership active=false by id
// @Tags         memberships
// @Accept       json
// @Produce      json
// @Param        id path string true "Membership ID"
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /memberships/{id} [patch]
func DeactivateMembership() fiber.Handler {
    return func(c *fiber.Ctx) error {
        idHex := c.Params("id")
        oid, err := bson.ObjectIDFromHex(idHex)
        if err != nil {
            return fiber.NewError(fiber.StatusBadRequest, "invalid id")
        }
        col := database.DB.Collection("memberships")
        _, err = col.UpdateOne(c.Context(), bson.M{"_id": oid}, bson.M{"$set": bson.M{"active": false}})
        if err != nil {
            return fiber.NewError(fiber.StatusInternalServerError, err.Error())
        }
        return c.JSON(fiber.Map{"_id": idHex, "active": false})
    }
}
