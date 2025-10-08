// authctx/user_id.go
package authctx

import (
	"gopkg.in/mgo.v2/bson"
	"github.com/gofiber/fiber/v2"
)

func UserIDFrom(c *fiber.Ctx) (bson.ObjectId, bool) {
	if v := c.Locals("user_id"); v != nil {
		if s, ok := v.(string); ok {
			if oid := bson.ObjectIdHex(s); oid.Valid() {
				return oid, true
			}
		}
	}
	return bson.ObjectId(""), false
}
