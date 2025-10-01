package dto

import "time"

// ========== SUPPORT STRUCTS ==========

// ใครเป็นคนโพสต์ (เช่น user หรือโพสต์ในนามตำแหน่ง)
type PostedAs struct {
	UserID      string `json:"user_id,omitempty"`
	PositionKey string `json:"position_key,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
}

// การมองเห็นของ Event
type Visibility struct {
	Access   string   `json:"access"`             // "public" | "org"
	Audience []string `json:"audience,omitempty"` // รายชื่อ org paths ที่เห็นได้
}

// ========== REQUEST ==========

// Incoming request payload when creating event
type EventRequest struct {
	NodeID          string      `json:"node_id"`
	Topic           string      `json:"topic"`
	Description     string      `json:"description"`
	MaxParticipation int        `json:"max_participation"`
	PostedAs        *PostedAs   `json:"posted_as,omitempty"`
	Visibility      *Visibility `json:"visibility,omitempty"`
	OrgOfContent    string      `json:"org_of_content,omitempty"`
	Status          string      `json:"status,omitempty"`

	Schedules []struct {
		Date        time.Time `json:"date"`
		TimeStart   time.Time `json:"time_start"`
		TimeEnd     time.Time `json:"time_end"`
		Location    string    `json:"location"`
		Description string    `json:"description"`
	} `json:"schedules"`
}

// ========== RESPONSE DTO ==========

// Reported schedule for response
type EventScheduleReport struct {
	Date      time.Time `json:"date"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// Response payload when returning created/queried event
type EventReport struct {
	EventID    string                `json:"event_id"`
	EventTopic string                `json:"event_topic"`
	Schedules  []EventScheduleReport `json:"schedules"`
}
