package dto

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"main-webbase/internal/models"
)

// Send all form structure question
type FormQuestionCreateDTO struct {
	Questions []QuestionItemDTO `json:"questions" validate:"required,dive"`
}

type QuestionItemDTO struct {
	QuestionText string `json:"question_text" validate:"required" example:"What is your name?"`
	Required     bool   `json:"required" example:"true"`
	OrderIndex   int    `json:"order_index" example:"1"`
}

// User submit response
type FormResponseSubmitDTO struct {
	Answers []AnswerSubmitItemDTO `json:"answers" validate:"required,dive"`
}

type AnswerSubmitItemDTO struct {
	QuestionID  string `json:"question_id" validate:"required"`
	AnswerValue string `json:"answer_value"`
	OrderIndex  int    `json:"order_index" example:"1"`
}

// To Get All User Answer with Question
type QuestionDTO struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type UserAnswersDTO struct {
	UserID    string   `json:"user_id"`
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Status    string   `json:"status"`
	Answers   []string `json:"answers"`
}

type FormMatrixResponseDTO struct {
	FormID    string           `json:"form_id"`
	Questions []QuestionDTO    `json:"questions"`
	Responses []UserAnswersDTO `json:"responses"`
}

type AggregateResponse struct {
	ResponseID  bson.ObjectID
	User        models.User
	Participant models.Event_participant
	Answers     []models.Event_form_answer
}

// Update Participant Status
type UpdateParticipantStatusDTO struct {
	UserID  string `json:"user_id" validate:"required" example:"66ffb4e71b64d7a993d53400"`
	EventID string `json:"event_id" validate:"required" example:"66ffb4e71b64d7a993d53401"`
	Status string `json:"status" validate:"required" example:"accept" enums:"accept,stall,reject"`
}