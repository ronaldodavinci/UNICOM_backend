package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"time"
)

type Membership struct {
	ID           bson.ObjectID   `bson:"_id,omitempty" json:"id"`
	UserID       bson.ObjectID   `bson:"user_id" json:"user_id"`
	OrgPath      string               `bson:"org_path" json:"org_path"`
	PositionKey  string               `bson:"position_key" json:"position_key"`
	Active       bool                 `bson:"active" json:"active"`
	OrgAncestors []string             `bson:"org_ancestors" json:"org_ancestors"`
	CreatedAt    time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time            `bson:"updated_at,omitempty" json:"updated_at,omitempty"`
}

type MembershipRequestDTO struct {
	UserID       string 		      `bson:"user_id" json:"user_id"`
	OrgPath      string               `bson:"org_path" json:"org_path"`
	PositionKey  string               `bson:"position_key" json:"position_key"`
	Active       bool                 `bson:"active" json:"active"`
}