package dto

// ===== Request =====
type CreatePostDTO struct {
	UserID      string   `json:"userId" validate:"required"`
	RoleID      string   `json:"roleId" validate:"required"`
	PostText    string   `json:"postText" validate:"required"`
	PictureUrl  []string `json:"pictureUrl"`
	VideoUrl    []string `json:"videoUrl"`
	CategoryIDs []string `json:"categoryIds"`
	RoleIDs     []string `json:"roleIds"`
}

// ===== Success Response =====
type PostResponse struct {
	UserID        string `json:"userId"        example:"66c6248b98c56c39f018e7d2"`
	RoleID        string `json:"roleId"        example:"66c6248b98c56c39f018e7d2"`
	PostText      string `json:"postText"      example:"สวัสดี KU!"`
	CreatedAt     string `json:"createdAt"     example:"2025-09-07T13:47:47Z"`
	UpdatedAt     string `json:"updatedAt"     example:"2025-09-07T13:47:47Z"`
	LikeCount     int    `json:"likeCount"     example:"0"`
	CommentCount  int    `json:"commentCount"  example:"0"`
	//PictureUrl    []string `json:"pictureUrl"    example:"['https://example.com/pic1.jpg','https://example.com/pic2.jpg']"`
	//VideoUrl      []string `json:"videoUrl"      example:"['https://example.com/vid1.mp4','https://example.com/vid2.mp4']"`
	CategoryIDs   []string `json:"categoryIds"   example:"['68bd8d30b98a8dce0eab0db6','68bd8d30b98a8dce0eab0db7']"`
	VisibleRoleIDs []string `json:"visibleRoleIds" example:"['66c6248b98c56c39f018e7d2','66c6248b98c56c39f018e7d3']"`
}

// ===== Error Response =====
type ErrorResponse struct {
	Message string `json:"message" example:"invalid body"`
}