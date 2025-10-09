// model/front_post.go (หรือเติมท้าย model_post.go)
package models

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
	ID        bson.ObjectID `bson:"_id" json:"_id"`
	UID       string        `bson:"uid" json:"uid"`
	Username  string        `bson:"username" json:"username"`
	Name      string        `bson:"name" json:"name"`
	Message   string        `bson:"message" json:"message"`
	Timestamp time.Time     `bson:"timestamp" json:"timestamp"`
	Likes     int           `bson:"likes" json:"likes"`
	PostedAs struct {
		OrgPath     string `bson:"org_path" json:"org_path"`
		PositionKey string `bson:"position_key" json:"position_key"`
	} `bson:"posted_as" json:"posted_as"`

	Tag string `bson:"tag" json:"tag"`

	Visibility struct {
		Access       string `bson:"access" json:"access"`
		OrgOfContent string `bson:"org_of_content" json:"org_of_content"`
	} `bson:"visibility" json:"visibility"`
}
