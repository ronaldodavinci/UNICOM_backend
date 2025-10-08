package dto

type UpdatePostFullDTO struct {
	PostText     string     `json:"postText" validate:"required"`
	PictureUrl   []string   `json:"pictureUrl"`
	VideoUrl     []string   `json:"videoUrl"`
	CategoryIDs  []string   `json:"categoryIds"`
	PostAs       PostAs     `json:"postAs" validate:"required"`
	Visibility   Visibility `json:"visibility" validate:"required"`
	OrgOfContent string     `json:"org_of_content"`
	Status       string     `json:"status"`
}
