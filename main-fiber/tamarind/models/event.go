package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Audience struct {
	OrgPath string `bson:"org_path" json:"org_path"`
	Scope   string `bson:"scope" json:"scope"`
}

type Visibility struct {
	Access   string     `bson:"access" json:"access"`
	Audience []Audience `bson:"audience" json:"audience"`
}

type PostedAs struct {
	OrgPath string `bson:"org_path" json:"org_path"`
}

type Event struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	NodeID           primitive.ObjectID `bson:"node_id" json:"node_id"`
	Topic            string             `bson:"topic" json:"topic"`
	Description      string             `bson:"description" json:"description"`
	MaxParticipation int                `bson:"max_participation" json:"max_participation"`
	PostedAs         PostedAs           `bson:"postedas" json:"postedas"`
	Visibility       Visibility         `bson:"visibility" json:"visibility"`
	OrgOfContent     string             `bson:"org_of_content" json:"org_of_content"`
	Status           string             `bson:"status" json:"status"`
	CreatedAt        time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time          `bson:"updated_at" json:"updated_at"`
}

type EventSchedule struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	EventID   primitive.ObjectID `bson:"event_id" json:"event_id"`
	StartTime time.Time          `bson:"start_time" json:"start_time"`
	EndTime   time.Time          `bson:"end_time" json:"end_time"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}