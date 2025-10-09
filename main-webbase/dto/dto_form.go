package dto

import (
	"main-webbase/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Initiate form
type FormCreateDTO struct {
	EventID 	string 	`json:"event_id" validate:"required"`
	OrgPath 	string 	`json:"org_path" validate:"required"`
}

// Send all form structure question
type FormQuestionCreateDTO struct {
	FormID    	string              	`json:"form_id" validate:"required"`
	Questions 	[]QuestionItemDTO     	`json:"questions" validate:"required,dive"`
}

type QuestionItemDTO struct {
	QuestionText 	string 	`json:"question_text" validate:"required"`
	Required     	bool   	`json:"required"`
	OrderIndex   	int    	`json:"order_index"`
}

// To publish form
type FormStatusUpdateDTO struct {
	FormID 	string 	`json:"form_id" validate:"required"`
	Status 	string 	`json:"status" validate:"required,oneof=Draft Published Inactive"`
}

// User submit response
type FormResponseSubmitDTO struct {
	FormID	string                    	`json:"form_id" validate:"required"`
	Answers	[]AnswerSubmitItemDTO  		`json:"answers" validate:"required,dive"`
}

type AnswerSubmitItemDTO struct {
	QuestionID	string		`json:"question_id" validate:"required"`
	AnswerValue	string 		`json:"answer_value"`
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
	FormID    string         `json:"form_id"`
	Questions []QuestionDTO  `json:"questions"`
	Responses []UserAnswersDTO `json:"responses"`
}

type AggregateResponse struct {
	ResponseID 		bson.ObjectID
	User 			models.User
	Participant 	models.Event_participant
	Answers 		[]models.Event_form_answer
}

// Update Participant Status
type UpdateParticipantStatusDTO struct {
	UserID 	string `json:"user_id"`
	EventID string `json:"event_id"`
	Status  string `json:"status"`
}