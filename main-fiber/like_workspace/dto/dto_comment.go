package dto

type CreateCommentReq struct {
	Text string `json:"text" validate:"required,min=1,max=2000"`
}

type UpdateCommentReq struct {
	Text string `json:"text" validate:"required,min=1,max=2000"`
}

type ListCommentsResp[T any] struct {
	Comments   []T     `json:"comments"`
	NextCursor *string `json:"next_cursor"`
	HasMore    bool    `json:"has_more"`
	TotalCount int64   `json:"total_count"`
}
