package model

type Post struct {
	ID        primitive.ObjectID `json:"id"         bson:"_id,omitempty"`
	UserID    primitive.ObjectID `json:"userId"     bson:"user_id"`
	RoleID    primitive.ObjectID `json:"roleId"     bson:"role_id"`
	PostText  string             `json:"postText"   bson:"post_text"`
	Picture   *string            `json:"pictureUrl,omitempty" bson:"picture_url,omitempty"`
	Video     *string            `json:"videoUrl,omitempty"   bson:"video_url,omitempty"`
	CreatedAt time.Time          `json:"createdAt"  bson:"created_at"`
	UpdatedAt time.Time          `json:"updatedAt"  bson:"updated_at"`
}
