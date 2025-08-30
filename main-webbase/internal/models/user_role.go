package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

type User_Role struct {
	ID     bson.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID bson.ObjectID `bson:"user_id" json:"user_id"`
	RoleID bson.ObjectID `bson:"role_id" json:"role_id"`
}

// Use to get request from api, Later change RoleID to bson.ObjectID
type UserRoleRequest struct {
	RoleID string `json:"role_id"`
}
