package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Role struct {
    ID             bson.ObjectID `bson:"_id,omitempty" json:"id"`
    RoleName       string             `bson:"role_name" json:"role_name"`
    RolePath       string             `bson:"role_path" json:"role_path"`
    PermBlog       bool               `bson:"perm_blog" json:"perm_blog"`
    PermEvent      bool               `bson:"perm_event" json:"perm_event"`
    PermComment    bool               `bson:"perm_comment" json:"perm_comment"`
    PermChildRole  bool               `bson:"perm_childrole" json:"perm_childrole"`
    PermSiblingRole bool              `bson:"perm_siblingrole" json:"perm_siblingrole"`
}

type User_Role struct {
	ID     bson.ObjectID `bson:"_id,omitempty" json:"id"`
    UserID bson.ObjectID `bson:"user_id" json:"user_id"`
    RoleID bson.ObjectID `bson:"role_id" json:"role_id"`
}