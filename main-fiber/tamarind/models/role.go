package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// Permission defines a single action key
type Permission struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Key      string             `bson:"key" json:"key"`
	Label    string             `bson:"label,omitempty" json:"label,omitempty"`
	Category string             `bson:"category,omitempty" json:"category,omitempty"`
}

// Role is a collection of permissions
type Role struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Label       string             `bson:"label" json:"label"`
	Permissions []string           `bson:"permissions" json:"permissions"`
}
