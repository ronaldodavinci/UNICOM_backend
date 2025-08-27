package dto

type CreatePostDTO struct {
	UserID      string   `json:"userId" validate:"required"`
	RoleID      string   `json:"roleId" validate:"required"`
	PostText    string   `json:"postText" validate:"required"`
	PictureUrl  *string  `json:"pictureUrl,omitempty"`   // ไม่ส่ง -> หายไป
	VideoUrl    *string  `json:"videoUrl,omitempty"`     // ไม่ส่ง -> หายไป
	CategoryIDs []string `json:"categoryIds,omitempty"` // ไม่ส่ง -> หายไป
}
