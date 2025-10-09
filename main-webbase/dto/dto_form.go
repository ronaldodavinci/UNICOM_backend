package dto

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
	UserID	string                    	`json:"user_id" validate:"required"`
	Answers	[]AnswerSubmitItemDTO  		`json:"answers" validate:"required,dive"`
}

type AnswerSubmitItemDTO struct {
	QuestionID	string		`json:"question_id" validate:"required"`
	AnswerValue	string 		`json:"answer_value"`
}

