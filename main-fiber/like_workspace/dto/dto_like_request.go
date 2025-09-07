package dto

type LikeRequestDTO struct {
	UserID    string `json:"userId" validate:"required"`
	PostID    string `json:"postId" validate:"required"`
}