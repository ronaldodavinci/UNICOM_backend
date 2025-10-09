package repository

import (
	"context"

	"main-webbase/database"
	"main-webbase/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func InitializeForm(ctx context.Context, form models.Event_form) error {
	_, err := database.DB.Collection("event_form").InsertOne(ctx, form)
	return err
}

func DeleteQuestionsByFormID(ctx context.Context, formID bson.ObjectID) error {
	_, err := database.DB.Collection("event_form_questions").DeleteMany(ctx, bson.M{"form_id": formID})
	return err
}

func InsertFormQuestions(ctx context.Context, questions []models.Event_form_question) error {
	var docs []interface{}
	for _, q := range questions {
		docs = append(docs, q)
	}

	_, err := database.DB.Collection("event_form_questions").InsertMany(ctx, docs)
	return err
}