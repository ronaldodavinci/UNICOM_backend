package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

type PostHashtag struct {
	ID     bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Date   string        `bson:"date"          json:"date"` // "YYYY-MM-DD"
	Tag    string        `bson:"tag"           json:"tag"`
	PostID bson.ObjectID `bson:"postId"        json:"postId"`
}
