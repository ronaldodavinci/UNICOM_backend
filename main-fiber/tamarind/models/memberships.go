package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

)

// Membership connects User ↔ OrgUnit ↔ Position
type Membership struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`
	OrgPath     string             `bson:"org_path" json:"org_path"`
	PositionKey string             `bson:"position_key" json:"position_key"`
	JoinedAt    *time.Time         `bson:"joined_at,omitempty" json:"joined_at,omitempty"`
	Active      *bool              `bson:"active,omitempty" json:"active,omitempty"`
}
