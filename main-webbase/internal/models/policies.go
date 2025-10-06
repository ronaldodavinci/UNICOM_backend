package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"time"
)

type Policy struct {
	ID          bson.ObjectID 		`bson:"_id,omitempty" json:"_id"`
	PositionKey string             	`bson:"position_key" json:"position_key"`
	Scope       string             	`bson:"scope" json:"scope"` // exact (control only the same orgpath) & subtree (control orgpath under this path)
	OrgPrefix 	string 				`bson:"org_prefix" json:"org_prefix"`
	Actions     []string           	`bson:"actions" json:"actions"`
	Enabled     bool               	`bson:"enabled" json:"enabled"`
	CreatedAt   time.Time          	`bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// Actions list
// "membership:assign"
// "organize:create"
// "event:create"