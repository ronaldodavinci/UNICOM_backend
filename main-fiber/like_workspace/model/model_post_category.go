package model

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

type PostCategory struct {
	ID         bson.ObjectID `json:"id"         bson:"_id,omitempty"`
	PostID     bson.ObjectID `json:"postId"     bson:"post_id"`
	CategoryID bson.ObjectID `json:"categoryId" bson:"category_id"`
	OrderIndex int           `json:"orderIndex" bson:"order_index"`
}
