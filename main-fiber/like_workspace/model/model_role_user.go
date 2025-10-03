package model

import "go.mongodb.org/mongo-driver/v2/bson"

type User struct {
	ID           bson.ObjectID `bson:"_id"`
	Firstname    string        `bson:"user_firstname"`
	Lastname     string        `bson:"user_lastname"`
	Status       string        `bson:"status"`
	ProfilePic   string        `bson:"profile_pic"`
	TypePerson   string        `bson:"type-person"`
	StudentID    string        `bson:"student_id"`
}

type Membership struct {
	ID           bson.ObjectID `bson:"_id"`
	NodeID	   	 bson.ObjectID `bson:"node_id"`
	UserID       bson.ObjectID `bson:"user_id"`
	OrgPath      string        `bson:"org_path"`       // e.g. "/faculty/eng/com"
	OrgAncestors []string      `bson:"org_ancestors"`  // ["/", "/faculty", "/faculty/eng"]
	PositionKey  string        `bson:"position_key"`   // e.g. "head"
	Active       bool          `bson:"active"`
	CreatedAt    any           `bson:"created_at"`
}

type OrgUnitNode struct {
	ID         bson.ObjectID `bson:"_id"`
	Path       string        `bson:"path"`            // "/faculty/eng/com"
	Ancestors  []string      `bson:"ancestors"`       // ["/", "/faculty", "/faculty/eng"]
	Depth      int           `bson:"depth"`
	ParentPath string        `bson:"parent_path"`
	Slug       string        `bson:"slug"`
	Status     string        `bson:"status"`          // "active"
	Visibility string        `bson:"visibility"`      // "public" | "private"
}

type Position struct {
	ID        bson.ObjectID `bson:"_id"`
	Key       string        `bson:"key"`              // "head"
	Status    string        `bson:"status"`
}
