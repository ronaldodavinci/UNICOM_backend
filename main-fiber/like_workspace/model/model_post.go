package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Post struct {
	ID        bson.ObjectID `json:"id"         bson:"_id,omitempty"`
	UserID    bson.ObjectID `json:"userId"     bson:"user_id"`
	RoleID    bson.ObjectID `json:"roleId"     bson:"role_id"`
	PostText  string        `json:"postText"   bson:"post_text"`
	Picture   *string       `json:"pictureUrl,omitempty" bson:"picture_url,omitempty"`
	Video     *string       `json:"videoUrl,omitempty"   bson:"video_url,omitempty"`
	CreatedAt time.Time     `json:"createdAt"  bson:"created_at"`
	UpdatedAt time.Time     `json:"updatedAt"  bson:"updated_at"`
}
