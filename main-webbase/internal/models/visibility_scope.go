package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"time"
)

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
