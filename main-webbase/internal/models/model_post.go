package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Post struct {
	ID           bson.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID       bson.ObjectID `json:"userId" bson:"user_id"`
	RolePathID   bson.ObjectID `json:"rolePathId"    bson:"node_id"`
	PositionID   bson.ObjectID `json:"positionId"    bson:"position_id"`
	Hashtag      []string      `json:"hashtag" bson:"hashtag"`
	Tags         string        `json:"tags" bson:"tags"`
	PostText     string        `json:"postText" bson:"post_text"`
	Media        []string      `bson:"media,omitempty" json:"media,omitempty"`
	CreatedAt    time.Time     `json:"createdAt" bson:"created_at"`
	UpdatedAt    time.Time     `json:"updatedAt" bson:"updated_at"`
	LikeCount    int           `json:"likeCount" bson:"like_count"`
	CommentCount int           `json:"CommentCount" bson:"comment_count"`
	Status       string        `json:"status" bson:"status"` // active, deleted
}

type QueryOptions struct {
	Roles      []string
	Categories []string
	Tags       []string
	AuthorIDs  []bson.ObjectID
	TextSearch string
	Limit      int64
	OnlyPublic *bool
	SinceID    bson.ObjectID
	UntilID    bson.ObjectID

	ViewerID       bson.ObjectID
	AllowedNodeIDs []bson.ObjectID
}
