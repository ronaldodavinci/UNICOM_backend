package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Media struct {
	ID        bson.ObjectID `bson:"_id" json:"id"`
	Type      string        `bson:"type" json:"type"`
	Path      string        `bson:"path" json:"path"`
	ThumbPath string        `bson:"thumb_path,omitempty" json:"thumbPath,omitempty"`
	Filename  string        `bson:"filename" json:"filename"`
	MIME      string        `bson:"mime" json:"mime"`
	Size      int64         `bson:"size" json:"size"`
	Width     int32         `bson:"width,omitempty" json:"width,omitempty"`
	Height    int32         `bson:"height,omitempty" json:"height,omitempty"`
	Duration  float64       `bson:"duration,omitempty" json:"duration,omitempty"`
	Order     int32         `bson:"order" json:"order"`

	// ฟิลด์ช่วยที่เราจะประกอบเองทีหลัง
	URL      string `bson:"-" json:"url"`
	ThumbURL string `bson:"-" json:"thumbUrl,omitempty"`
}

type Post struct {
	ID           bson.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID       bson.ObjectID `json:"userId" bson:"user_id"`
	RolePathID   bson.ObjectID `json:"rolePath" bson:"role_path"`
	PositionID   bson.ObjectID `json:"positionKey" bson:"position_key"`
	Hashtag      []string      `json:"hashtag" bson:"hashtag"`
	Tags         string        `json:"tags" bson:"tags"`
	PostText     string        `json:"postText" bson:"post_text"`
	CreatedAt    time.Time     `json:"createdAt" bson:"created_at"`
	UpdatedAt    time.Time     `json:"updatedAt" bson:"updated_at"`
	LikeCount    int           `json:"likeCount" bson:"like_count"`
	CommentCount int           `json:"CommmentCount" bson:"comment_count"`
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
