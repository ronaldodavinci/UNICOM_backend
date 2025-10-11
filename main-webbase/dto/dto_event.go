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
	OrgOfContent     string             `json:"org_of_content,omitempty"`
	Status           string             `json:"status,omitempty"`
	Have_form        bool               `json:"have_form,omitempty"`

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

// Event Detail
type EventDetail struct {
	EventID              string             `json:"event_id"`
	FormID               string             `json:"form_id,omitempty"`
	OrgPath              string             `json:"orgpath"`
	Topic                string             `json:"topic"`
	Description          string             `json:"description"`
	MaxParticipation     int                `json:"max_participation"`
	CurrentParticipation int                `json:"current_participation"`
	PostedAs             *models.PostedAs   `json:"posted_as,omitempty"`
	Visibility           *models.Visibility `json:"visibility,omitempty"`
	Status               string             `json:"status,omitempty"`
	Have_form            bool               `json:"have_form,omitempty"`

	Schedules []models.EventSchedule `json:"schedules"`
}
