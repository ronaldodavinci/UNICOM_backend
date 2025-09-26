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

func UserRolesFrom(c *fiber.Ctx) []string {
	// สมมติ middleware ใส่ roles มาใน Locals("roles") เป็น []string
	if v := c.Locals("roles"); v != nil {
		if roles, ok := v.([]string); ok {
			return roles
		}
		// กรณี JWT decode แล้วได้เป็น []interface{} ก็แปลง
		if any, ok := v.([]interface{}); ok {
			out := make([]string, 0, len(any))
			for _, x := range any {
				if s, ok := x.(string); ok {
					out = append(out, s)
				}
			}
			return out
		}
	}
	return nil
}
