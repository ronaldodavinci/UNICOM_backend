// model/front_post.go (หรือเติมท้าย model_post.go)
package model

import (
	"time"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type FrontPostedAs struct {
	OrgPath     string `json:"org_path"     bson:"org_path"`
	PositionKey string `json:"position_key" bson:"position_key"`
}

type FrontVisibility struct {
	Access       string `json:"access"         bson:"access"`          // "public" | "role"
	OrgOfContent string `json:"org_of_content" bson:"org_of_content"`  // role_path ของผู้เขียน
}

type FrontPost struct {
	ID        bson.ObjectID  `json:"_id"       bson:"_id"`
	UID       string         `json:"uid"       bson:"uid"`
	Name      string         `json:"name"      bson:"name"`
	Username  string         `json:"username"  bson:"username"`
	Message   string         `json:"message"   bson:"message"`
	Timestamp time.Time      `json:"timestamp" bson:"timestamp"`
	Likes     int            `json:"likes"     bson:"likes"`
	LikedBy   []string       `json:"likedBy"   bson:"likedBy"`
	PostedAs  FrontPostedAs  `json:"posted_as" bson:"posted_as"`
	Tag       string         `json:"tag"       bson:"tag"`
	Visibility FrontVisibility `json:"visibility" bson:"visibility"`
}
