package services
import (
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func FetchUserID(c *fiber.Ctx) (bson.ObjectID, bool) {
	if v := c.Locals("user_id"); v != nil {
		if s, ok := v.(string); ok {
			if oid, err := bson.ObjectIDFromHex(s); err == nil {
				return oid, true
			}
		}
	}
	return bson.NilObjectID, false
}
