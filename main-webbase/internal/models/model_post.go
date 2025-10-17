package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Post struct {
	ID           bson.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID       bson.ObjectID `json:"userId" bson:"user_id"`
	//Username  	 string        `bson:"username" json:"username"`
	Name      	 string        `bson:"name" json:"name"`
	RolePathID   bson.ObjectID `json:"rolePathId"    bson:"node_id"`
	PositionID   bson.ObjectID `json:"positionId"    bson:"position_id"`
	Hashtag      []string      `json:"hashtag" bson:"hashtag"`
	Tags         string        `json:"tags" bson:"tag"`
	Category     []string      `json:"category" bson:"category"`
	PostText     string        `json:"postText" bson:"post_text"`
	Media        []string      `bson:"media,omitempty" json:"media,omitempty"`
	LikeCount    int           `json:"likeCount" bson:"like_count"`
	CommentCount int           `json:"CommentCount" bson:"comment_count"`
	CreatedAt    time.Time     `json:"createdAt" bson:"created_at"`
	UpdatedAt    time.Time     `json:"updatedAt" bson:"updated_at"`
	Status       string        `json:"status" bson:"status"` // active, deleted
	Visibility   string        `json:"visibility" bson:"visibility"`
	Isliked      bool          `json:"is_liked" bson:"is_liked"`
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
