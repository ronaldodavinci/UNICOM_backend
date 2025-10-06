package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"time"
)

type Constraints struct {
    ExclusivePerOrg bool `bson:"exclusive_per_org" json:"exclusive_per_org"`
}

type Scope struct {
    OrgPath string `bson:"org_path" json:"org_path"`
    Inherit bool   `bson:"inherit" json:"inherit"`
}

type Position struct {
    ID          bson.ObjectID `bson:"_id,omitempty" json:"_id"`
    Key         string             `bson:"key" json:"key"`
    Constraints Constraints        `bson:"constraints" json:"constraints"`
    Display     map[string]string  `bson:"display" json:"display"`
    Rank        int                `bson:"rank" json:"rank"`
    Scope       Scope              `bson:"scope" json:"scope"`
    Status      string             `bson:"status" json:"status"`
    CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
    UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
}
