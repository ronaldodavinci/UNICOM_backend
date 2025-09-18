package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type ModelPost struct {
	ID           bson.ObjectID `json:"id"         bson:"_id,omitempty"`
	UserID       bson.ObjectID `json:"userId"     bson:"user_id"`
	RoleID       bson.ObjectID `json:"roleId"     bson:"role_id"`
	PostText     string        `json:"postText"   bson:"post_text"`
	CreatedAt    time.Time     `json:"createdAt"  bson:"created_at"`
	UpdatedAt    time.Time     `json:"updatedAt"  bson:"updated_at"`
	LikeCount    int           `json:"likeCount"  bson:"like_count"`
	CommentCount int           `json:"CommmentCount"  bson:"comment_count"`
}