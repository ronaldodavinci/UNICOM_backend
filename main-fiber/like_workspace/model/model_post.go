package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Post struct {
	ID           bson.ObjectID `json:"id"         bson:"_id,omitempty"`
	UserID       bson.ObjectID `json:"userId"     bson:"user_id"`
	RolePathID   bson.ObjectID `json:"rolePath"     bson:"role_path"`
	PositionID   bson.ObjectID `json:"positionKey"     bson:"position_key"`
	Hashtag 	 []string        `json:"hashtag"    bson:"hashtag"`
	Tags         string        `json:"tags"       bson:"tags"`
	PostText     string        `json:"postText"   bson:"post_text"`
	CreatedAt    time.Time     `json:"createdAt"  bson:"created_at"`
	UpdatedAt    time.Time     `json:"updatedAt"  bson:"updated_at"`
	LikeCount    int           `json:"likeCount"  bson:"like_count"`
	CommentCount int           `json:"CommmentCount"  bson:"comment_count"`
}

type QueryOptions struct {
	Roles      []string
	Categories []string
	Tags       []string
	AuthorIDs  []bson.ObjectID
	TextSearch string
	Limit      int64
	SinceID    bson.ObjectID
	UntilID    bson.ObjectID
}