package cursor

import (
	"encoding/base64"
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Cursor (ใช้ createdAt + _id)
type Cursor struct {
	CreatedAt int64  `json:"createdAt"`
	ID        string `json:"id"`
}

func EncodeCursor(t time.Time, id bson.ObjectID) string {
	b, _ := json.Marshal(Cursor{
		CreatedAt: t.UnixMilli(),
		ID:        id.Hex(),
	})
	return base64.StdEncoding.EncodeToString(b)
}

func DecodeCursor(s string) (time.Time, bson.ObjectID, error) {
	raw, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return time.Time{}, bson.NilObjectID, err
	}

	var p Cursor
	if err := json.Unmarshal(raw, &p); err != nil {
		return time.Time{}, bson.NilObjectID, err
	}

	oid, err := bson.ObjectIDFromHex(p.ID)
	if err != nil {
		return time.Time{}, bson.NilObjectID, err
	}

	return time.UnixMilli(p.CreatedAt).UTC(), oid, nil
}
