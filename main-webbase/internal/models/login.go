package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ====== Permission ======
type Permission struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Key      string             `bson:"key" json:"key"`
	Label    string             `bson:"label,omitempty" json:"label,omitempty"`
	Category string             `bson:"category,omitempty" json:"category,omitempty"`
}

// ====== Role ======

// ====== User with role expansion ======
type UserWithRoleDetails struct {
	User        `bson:",inline"`
	RoleDetails []Role `bson:"roleDetails" json:"roleDetails"`
}
