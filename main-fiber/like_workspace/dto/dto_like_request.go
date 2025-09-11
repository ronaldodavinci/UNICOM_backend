package dto

type LikeRequestDTO struct {
	UserID     string `json:"userId"`
	TargetID   string `json:"targetId"`
	TargetType string `json:"targetType"` // "post" | "comment"
}

type LikeErrorResponse struct {
	Message string `json:"message"`
}