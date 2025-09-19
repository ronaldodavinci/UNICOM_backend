package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"time"
)

type Event struct {
	ID               bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Node_ID          bson.ObjectID `bson:"node_id" json:"node_id"`
	Topic            string        `bson:"topic" json:"topic"`
	Description      string        `bson:"description" json:"description"`
	MaxParticipation int           `bson:"max_participation" json:"max_participation"`
	CreatedAt        *time.Time    `bson:"created_at,omitempty" json:"created_at,omitempty"`
	UpdatedAt        *time.Time    `bson:"updated_at,omitempty" json:"updated_at,omitempty"`
}

type EventSchedule struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Event_ID    bson.ObjectID `bson:"event_id" json:"event_id"`
	Date        time.Time     `bson:"date" json:"date"`
	Time_start  time.Time     `bson:"time_start" json:"time_start"`
	Time_end    time.Time     `bson:"time_end" json:"time_end"`
	Location    string        `bson:"location" json:"location"`
	Description string        `bson:"description" json:"description"`
}
