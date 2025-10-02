package models

import (
    "time"

    "go.mongodb.org/mongo-driver/bson/primitive"
)

type PolicyWhere struct {
    OrgPrefix string `bson:"org_prefix" json:"org_prefix"`
}

type Policy struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	PositionKey string             `bson:"position_key" json:"position_key"`
	Scope       string             `bson:"scope" json:"scope"`
	Where       PolicyWhere        `bson:"where" json:"where"`
	Actions     []string           `bson:"actions" json:"actions"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	Effect      string             `bson:"effect" json:"effect"`
	Enabled     bool               `bson:"enabled" json:"enabled"`
}