package models

import (
    "time"
    "go.mongodb.org/mongo-driver/v2/bson/primitive"
)

// Event main document
type Event struct {
    ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    NodeID           primitive.ObjectID `bson:"node_id" json:"node_id"`
    Topic            string             `bson:"topic" json:"topic"`
    Description      string             `bson:"description" json:"description"`
    MaxParticipation int                `bson:"max_participation" json:"max_participation"`
    OrgOfContent     string             `bson:"org_of_content,omitempty" json:"org_of_content,omitempty"`
    Status           string             `bson:"status,omitempty" json:"status,omitempty"`
    CreatedAt        *time.Time         `bson:"created_at,omitempty" json:"created_at,omitempty"`
    UpdatedAt        *time.Time         `bson:"updated_at,omitempty" json:"updated_at,omitempty"`
}

// EventSchedule stores schedule of event
type EventSchedule struct {
    ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    EventID     primitive.ObjectID `bson:"event_id" json:"event_id"`
    Date        time.Time          `bson:"date" json:"date"`
    TimeStart   time.Time          `bson:"time_start" json:"time_start"`
    TimeEnd     time.Time          `bson:"time_end" json:"time_end"`
    Location    *string            `bson:"location,omitempty" json:"location,omitempty"`
    Description *string            `bson:"description,omitempty" json:"description,omitempty"`
}
