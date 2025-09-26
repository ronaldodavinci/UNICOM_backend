package dto

import (
	"main-webbase/internal/models"
	"time"
)

// Send Everything about event
type EventRequestDTO struct {
	NodeID           string             `json:"node_id"`
	Topic            string             `json:"topic"`
	Description      string             `json:"description"`
	MaxParticipation int                `json:"max_participation"`
	PostedAs         *models.PostedAs   `json:"posted_as,omitempty"`
	Visibility       *models.Visibility `json:"visibility,omitempty"`
	OrgOfContent     string             `bson:"org_of_content,omitempty" json:"org_of_content,omitempty"`
	Status           string             `bson:"status,omitempty" json:"status,omitempty"`

	Schedules []struct {
		Date        time.Time `json:"date"`
		Time_start  time.Time `json:"time_start"`
		Time_end    time.Time `json:"time_end"`
		Location    string    `json:"location"`
		Description string    `json:"description"`
	} `json:"schedules"`
}

// Report DTO
type EventScheduleReport struct {
	Date      time.Time `json:"date"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

type EventReport struct {
	EventID    string                `json:"event_id"`
	EventTopic string                `json:"event_topic"`
	Schedules  []EventScheduleReport `json:"schedules"`
}
