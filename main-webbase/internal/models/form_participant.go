package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"time"
)

type Event_form struct { 
	ID 			bson.ObjectID 	`bson:"_id,omitempty" json:"id"`
	Event_ID 	bson.ObjectID 	`bson:"event_id" json:"event_id"` 
	OrgPath 	string 			`bson:"org_path" json:"org_path"` 
	Status 		string 			`bson:"status" json:"status"` // Published / Draft / Inactive 
	CreatedAt 	*time.Time 		`bson:"created_at,omitempty" json:"created_at,omitempty"` 
	UpdatedAt 	*time.Time 		`bson:"updated_at,omitempty" json:"updated_at,omitempty"` 
}

type Event_participant struct {
	ID          bson.ObjectID 	`bson:"_id,omitempty" json:"id"`
	Event_ID    bson.ObjectID 	`bson:"event_id" json:"event_id"`
	User_ID     bson.ObjectID 	`bson:"user_id" json:"user_id"`
	Response_ID bson.ObjectID 	`bson:"response_id" json:"response_id"`
	Status 		string 		   	`bson:"status" json:"status"` // Accept / Stall / Declined
	Role        string        	`bson:"role" json:"role"`// Organizer / Participant
	CreatedAt   *time.Time    	`bson:"created_at,omitempty" json:"created_at,omitempty"`
}

type Event_response struct {
	ID       bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Form_ID  bson.ObjectID `bson:"form_id" json:"form_id"`
	User_ID  bson.ObjectID `bson:"user_id" json:"user_id"`
	SubmitAt *time.Time    `bson:"submitted_at,omitempty" json:"submitted_at,omitempty"`
}

type Event_form_question struct {
	ID            bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Form_ID       bson.ObjectID `bson:"form_id" json:"form_id"`
	Question_text string        `bson:"question_text" json:"question_text"`
	Required      bool     		`bson:"required" json:"required"` // Optional / Required
	OrderIndex    int			`bson:"order_index" json:"order_index"`
	CreatedAt   *time.Time    	`bson:"created_at,omitempty" json:"created_at,omitempty"`
}

type Event_form_answer struct {
	ID          	bson.ObjectID 	`bson:"_id,omitempty" json:"id"`
	Question_ID 	bson.ObjectID 	`bson:"question_id" json:"question_id"`
	Response_ID 	bson.ObjectID 	`bson:"response_id" json:"response_id"`
	Answer_value	string 			`bson:"answer_value,omitempty" json:"answer_value,omitempty"`
	CreatedAt   	*time.Time    	`bson:"created_at,omitempty" json:"created_at,omitempty"`
}
