package dto

// Object in reequest body
type PostAs struct {
	OrgPath     string `bson:"org_path,omitempty"     json:"org_path,omitempty"`
	PositionKey string `bson:"position_key,omitempty" json:"position_key,omitempty"`
	Tag         string `bson:"label,omitempty"        json:"label,omitempty"` // e.g., "Head • SMO"
}

type Visibility struct {
	Access   string   `json:"access"`
	Audience []string `json:"audience"`
}

// ===== Request =====
type CreatePostDTO struct {
	PostText string   `json:"postText" validate:"required"`
	Media    []string `json:"media,omitempty"`

	CategoryIDs []string `json:"categoryIds"`

	PostAs       PostAs     `json:"postAs" validate:"required"`     // เดิมคือ rolePath
	Visibility   Visibility `json:"visibility" validate:"required"` // เดิมคือ roleIds
	OrgOfContent string     `json:"org_of_content"`
}

/* { request example
  "_id": {
    "$oid": "68be742243c7f21d8421a0e7"
  },
  "uid": "u_jy",
  "name": "JY",
  "username": "jy_smo",
  "message": "SMO is hosting an ENG night this Friday! 🏮",
  "timestamp": {
    "$date": "2025-09-08T06:13:54.392Z"
  },
  "likes": 2,
  "likedBy": [
    "u_002",
    "u_005"
  ],
  "posted_as": {
    "org_path": "/faculty/eng/smo",
    "position_key": "head",
    "tag": "Head • SMO"
  },
  "visibility": {
    "access": "org",
    "audience": [
      {
        "org_path": "/faculty/eng",
        "scope": "subtree"
      }
    ]
  },
  "org_of_content": "/faculty/eng/smo"
} */

// ===== Success Response =====
type PostResponse struct {
	UserID       string     `json:"userId"        example:"66c6248b98c56c39f018e7d2"`
	Name         string     `json:"name"          example:"JY"`
	Username     string     `json:"username"      example:"jy_smo"`
	PostText     string     `json:"postText"      example:"สวัสดี KU!"`
	Media        []string   `json:"media,omitempty" example:"['ttp://45.144.166.252:46602/uploads/cat.png']"`
	Hashtag      []string   `json:"hashtag" bson:"hashtag"`
	LikeCount    int        `json:"likeCount"     example:"0"`
	CommentCount int        `json:"commentCount"  example:"0"`
	PostAs       PostAs     `json:"postAs"` // which role is posting this
	CategoryIDs  []string   `json:"categoryIds"   example:"['68bd8d30b98a8dce0eab0db6','68bd8d30b98a8dce0eab0db7']"`
	Visibility   Visibility `json:"visibility"`     // which roles can see this post, array of roleIds
	OrgOfContent string     `json:"org_of_content"` // this content belongs to which org (org_path)
	CreatedAt    string     `json:"createdAt"     example:"2025-09-07T13:47:47Z"`
	UpdatedAt    string     `json:"updatedAt"     example:"2025-09-07T13:47:47Z"`
	Status       string     `json:"status" example:"active"`
}

// ===== Error Response =====
type ErrorResponse struct {
	Message string `json:"message" example:"invalid body"`
}
