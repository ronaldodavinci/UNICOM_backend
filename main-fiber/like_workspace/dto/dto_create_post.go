package dto

type CreatePostDTO struct {
	UserID      string   `json:"userId" validate:"required"`
	RoleID      string   `json:"roleId" validate:"required"`
	PostText    string   `json:"postText" validate:"required"`
	PictureUrl  []string `json:"pictureUrl"`
	VideoUrl    []string `json:"videoUrl"`
	CategoryIDs []string `json:"categoryIds"`
	RoleIDs     []string `json:"roleIds"`
}
