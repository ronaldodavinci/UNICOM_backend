package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Comment struct {
	ID        bson.ObjectID `json:"id"        bson:"_id,omitempty"`
	PostID    bson.ObjectID `json:"postId"    bson:"post_id"`
	UserID    bson.ObjectID `json:"userId"    bson:"user_id"`
	Text      string        `json:"text"      bson:"text"`
	CreatedAt time.Time     `json:"createdAt" bson:"created_at"`
	UpdatedAt time.Time     `json:"updatedAt" bson:"updated_at"`
	LikeCount int           `json:"likeCount" bson:"like_count"`
}
