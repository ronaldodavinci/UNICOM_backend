package dto

import (
	"time"
)

// -- Request --
// สำหรับ POST /events/:eventId/qa
// ถามคำถามเกี่ยวกับ Event
type CreateQADTO struct {
	QuestionText string `json:"questionText" validate:"required,min=1,max=2000"`
}

// PATCH /qa/:qaId/answer
// ตอบคำถามเกี่ยวกับ Event
type AnswerQADTO struct {
	AnswerText string `json:"answerText" validate:"required,min=1,max=2000"`
}


// -- Response --
type EventQAResponse struct {
	ID                string     `json:"id"`
	EventID           string     `json:"eventId"`
	QuestionerID      string     `json:"questionerId"`
	AnswererID        string     `json:"answererId"`
	QuestionText      string     `json:"questionText"`
	QuestionCreatedAt time.Time  `json:"questionCreatedAt"`
	AnswerText        *string    `json:"answerText,omitempty"`
	AnswerCreatedAt   *time.Time `json:"answerCreatedAt,omitempty"`
	Status            string     `json:"status"`
}
