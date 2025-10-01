package dto

import "time"

// Post creation request
type PostRequest struct {
	NodeID   string `json:"node_id"`
	PostText string `json:"post_text"`
}

// Post response
type PostResponse struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	NodeID       string    `json:"node_id"`
	PostText     string    `json:"post_text"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	LikeCount    int       `json:"like_count"`
	CommentCount int       `json:"comment_count"`
}
