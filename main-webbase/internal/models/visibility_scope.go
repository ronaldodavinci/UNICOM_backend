package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"time"
)

type AudienceItem struct {
	OrgPath string `bson:"org_path" json:"org_path"`
	Scope   string `bson:"scope"    json:"scope"` // "exact" | "subtree"
}

type Visibility struct {
	Access           string         `bson:"access,omitempty"            json:"access,omitempty"`
	Audience         []AudienceItem `bson:"audience,omitempty"          json:"audience,omitempty"`
	IncludePositions []string       `bson:"include_positions,omitempty" json:"include_positions,omitempty"`
	ExcludePositions []string       `bson:"exclude_positions,omitempty" json:"exclude_positions,omitempty"`
	AllowUserIDs     []string       `bson:"allow_user_ids,omitempty"    json:"allow_user_ids,omitempty"`
	DenyUserIDs      []string       `bson:"deny_user_ids,omitempty"     json:"deny_user_ids,omitempty"`
}

type PostedAs struct {
	OrgPath     string `bson:"org_path,omitempty"     json:"org_path,omitempty"`
	PositionKey string `bson:"position_key,omitempty" json:"position_key,omitempty"`
	Label       string `bson:"label,omitempty"        json:"label,omitempty"`
}

type MembershipDoc struct {
	ID          bson.ObjectID `bson:"_id,omitempty"    json:"_id"`
	UserID      bson.ObjectID `bson:"user_id"          json:"user_id"`
	OrgPath     string        `bson:"org_path"         json:"org_path"`
	PositionKey string        `bson:"position_key"     json:"position_key"`
	Active      bool          `bson:"active"           json:"active"` // boolean flag
	JoinedAt    *time.Time    `bson:"joined_at,omitempty" json:"joined_at,omitempty"`
	EndedAt     *time.Time    `bson:"ended_at,omitempty"  json:"ended_at,omitempty"`
	CreatedAt   time.Time     `bson:"created_at"       json:"created_at"`
	UpdatedAt   time.Time     `bson:"updated_at"       json:"updated_at"`
}
