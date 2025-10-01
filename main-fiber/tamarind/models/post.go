package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Post document
type Post struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`
	NodeID      primitive.ObjectID `bson:"node_id" json:"node_id"`
	PostText    string             `bson:"post_text" json:"post_text"`
	CreatedAt   *time.Time         `bson:"created_at,omitempty" json:"created_at,omitempty"`
	UpdatedAt   *time.Time         `bson:"updated_at,omitempty" json:"updated_at,omitempty"`
	LikeCount   int                `bson:"like_count" json:"like_count"`
	CommentCount int               `bson:"comment_count" json:"comment_count"`
}

// Comment document
type Comment struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	PostID    primitive.ObjectID `bson:"post_id" json:"post_id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Text      string             `bson:"comment_text" json:"comment_text"`
	CreatedAt *time.Time         `bson:"created_at,omitempty" json:"created_at,omitempty"`
	UpdatedAt *time.Time         `bson:"updated_at,omitempty" json:"updated_at,omitempty"`
}

// Like document
type Like struct {
	ID     primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	PostID primitive.ObjectID `bson:"post_id" json:"post_id"`
	UserID primitive.ObjectID `bson:"user_id" json:"user_id"`
}
