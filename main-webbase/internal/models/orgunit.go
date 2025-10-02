package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

// OrgUnit document stored in "org_units"
type OrgUnit struct {
	ID        bson.ObjectID     `bson:"_id,omitempty" json:"_id,omitempty"`
	OrgPath   string            `bson:"org_path"`
	Parent    string            `bson:"parent_path,omitempty"`
	Type      string            `bson:"type,omitempty"`
	Label     map[string]string `bson:"label,omitempty"`
	ShortName string            `bson:"short_name,omitempty"`
	Sort      int               `bson:"sort,omitempty"`
}
