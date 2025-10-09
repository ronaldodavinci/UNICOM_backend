package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type EventQA struct {
	ID                bson.ObjectID `bson:"_id"                  json:"id"`
	EventID           bson.ObjectID `bson:"event_id"             json:"eventId"`
	QuestionerID      bson.ObjectID `bson:"questioner_id"        json:"questionerId"`
	AnswererID        bson.ObjectID `bson:"answerer_id"          json:"answererId"` // เจ้าของ event

	QuestionText      string        `bson:"question_text"        json:"questionText"`
	QuestionCreatedAt time.Time     `bson:"question_created_at"  json:"questionCreatedAt"`

	AnswerText        *string       `bson:"answer_text,omitempty"       json:"answerText,omitempty"`
	AnswerCreatedAt   *time.Time    `bson:"answer_created_at,omitempty" json:"answerCreatedAt,omitempty"`

	Status            string        `bson:"status"               json:"status"` // "pending" | "answered" | "deleted"
}
