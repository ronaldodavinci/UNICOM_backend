package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// Position main document
type Position struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	Scope     string             `bson:"scope" json:"scope"`
	Status    string             `bson:"status" json:"status"`
	CreatedAt *time.Time         `bson:"created_at,omitempty" json:"created_at,omitempty"`
	UpdatedAt *time.Time         `bson:"updated_at,omitempty" json:"updated_at,omitempty"`
}
