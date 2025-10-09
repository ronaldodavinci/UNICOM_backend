package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

type PostRoleVisibility struct {
	ID     bson.ObjectID `json:"id"         bson:"_id,omitempty"`
	PostID bson.ObjectID `json:"postId"     bson:"post_id"`
	NodeID *bson.ObjectID `json:"nodeID"     bson:"node_id"`
}
