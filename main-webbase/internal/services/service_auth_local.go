package services

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func UserIDFrom(c *fiber.Ctx) (bson.ObjectID, error) {
	v := c.Locals("user_id")
	if v == nil {
		return bson.NilObjectID, fmt.Errorf("no user in context")
	}
	s, ok := v.(string)
	if !ok {
		return bson.NilObjectID, fmt.Errorf("user_id is not string")
	}
	oid, err := bson.ObjectIDFromHex(s)
	if err != nil {
		return bson.NilObjectID, fmt.Errorf("invalid objectID")
	}
	return oid, nil
}