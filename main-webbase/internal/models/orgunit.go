package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"time"
)

type OrgUnitNode struct {
	ID           bson.ObjectID   	`bson:"_id,omitempty" json:"id"`
	OrgPath   	 string            	`bson:"org_path" json:"org_path"`
	ParentPath   string            	`bson:"parent_path,omitempty" json:"parent_path"`
	Ancestors  	 []string          	`bson:"ancestors"   json:"ancestors"`
	Depth      	 int               	`bson:"depth"       json:"depth"`
	Name       	 string 			`bson:"name"        json:"name"`
	ShortName    string 			`bson:"shortname"   json:"shortname"`
	Slug      	 string             `bson:"slug,omitempty" json:"slug,omitempty"`
	Type       	 string            	`bson:"type"        json:"type"`
	Status     	 string            	`bson:"status"      json:"status"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at,omitempty" json:"updated_at,omitempty"`
}