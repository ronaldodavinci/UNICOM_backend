package dto

import "main-webbase/internal/models"

type CreateCommentReq struct {
	Text string `json:"text" validate:"required,min=1,max=2000"`
}

type UpdateCommentReq struct {
	Text string `json:"text" validate:"required,min=1,max=2000"`
}

type ListCommentsResp struct {
	Comments   []models.Comment `json:"comments"`
	NextCursor *string          `json:"next_cursor"`
	HasMore    bool             `json:"has_more"`
}
