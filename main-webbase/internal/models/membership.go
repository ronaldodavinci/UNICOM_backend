package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"time"
)

// Membership connects User ↔ OrgUnit ↔ Position
type Membership struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	UserID      bson.ObjectID `bson:"user_id" json:"user_id"`
	OrgPath     string        `bson:"org_path" json:"org_path"`
	PositionKey string        `bson:"position_key" json:"position_key"`
	JoinedAt    *time.Time    `bson:"joined_at,omitempty" json:"joined_at,omitempty"`
	Active      *bool         `bson:"active,omitempty" json:"active,omitempty"`
}
