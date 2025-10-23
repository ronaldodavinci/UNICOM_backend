package dto

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type CreateCommentReq struct {
	Text string `json:"text" validate:"required,min=1,max=2000"`
}

type UpdateCommentReq struct {
	Text string `json:"text" validate:"required,min=1,max=2000"`
}

// CommentResp is the API response for a comment with filtered text.
type CommentResp struct {
	ID           bson.ObjectID `json:"id"`
	PostID       bson.ObjectID `json:"postId"`
	UserID       bson.ObjectID `json:"userId"`
	CensoredText string        `json:"text"`
	CreatedAt    time.Time     `json:"createdAt"`
	UpdatedAt    time.Time     `json:"updatedAt"`
	LikeCount    int           `json:"likeCount"`
}

type ListCommentsResp struct {
	Comments   []CommentResp `json:"comments"`
	NextCursor *string       `json:"next_cursor"`
	HasMore    bool          `json:"has_more"`
	IsLiked    bool          `json:"isLiked"`
}
