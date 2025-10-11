package repository

import (
	"context"
	"fmt"

	"main-webbase/database"
	"main-webbase/dto"
	"main-webbase/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
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

func AddParticipant(ctx context.Context, participant models.Event_participant) error {
	_, err := database.DB.Collection("event_participant").InsertOne(ctx, participant)
	return err
}

func AddListParticipant(ctx context.Context, participants []models.Event_participant) error {
	if len(participants) == 0 {
		return nil
	}

	docs := make([]interface{}, len(participants))
	for i, p := range participants {
		docs[i] = p
	}

	_, err := database.DB.Collection("event_participant").InsertMany(ctx, docs)
	return err
}

func FindFormByID(ctx context.Context, formID string) (*models.Event_form, error) {
	objectID, err := bson.ObjectIDFromHex(formID)
	if err != nil {
		return nil, fmt.Errorf("invalid form ID: %w", err)
	}

	var form models.Event_form
	err = database.DB.Collection("event_form").FindOne(ctx, bson.M{"_id": objectID}).Decode(&form)
	if err != nil {
		return nil, err
	}

	return &form, nil
}

func FindQuestionsByFormID(ctx context.Context, formID bson.ObjectID) ([]models.Event_form_question, error) {
	collection := database.DB.Collection("event_form_questions")

	// Sort by order_index ascending
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "order_index", Value: 1}})

	cursor, err := collection.Find(ctx, bson.M{"form_id": formID}, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var questions []models.Event_form_question
	if err := cursor.All(ctx, &questions); err != nil {
		return nil, err
	}

	return questions, nil
}

func SubmitResponse(ctx context.Context, response models.Event_response) error {
	_, err := database.DB.Collection("event_form_response").InsertOne(ctx, response)
	return err
}

func InsertAnswers(ctx context.Context, docs []interface{}) error {
	_, err := database.DB.Collection("event_form_answer").InsertMany(ctx, docs)
	return err
}

func HasUserSubmittedResponse(ctx context.Context, formID, userID string) (bool, error) {
	FormID, err := bson.ObjectIDFromHex(formID)
	if err != nil {
		return false, err
	}

	UserID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return false, err
	}

	count, err := database.DB.Collection("event_form_response").CountDocuments(ctx, bson.M{"form_id": FormID, "user_id": UserID})
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func AggregateUserResponse(ctx context.Context, formID bson.ObjectID) ([]dto.AggregateResponse, error) {
	collection := database.DB.Collection("event_form_response")

	// MongoDB aggregation pipeline
	pipeline := mongo.Pipeline{
		// 1. Match responses for this form
		{{Key: "$match", Value: bson.M{"form_id": formID}}},
		// 2. Lookup answers
		{{
			Key: "$lookup",
			Value: bson.M{
				"from":         "event_form_answer",
				"localField":   "_id",
				"foreignField": "response_id",
				"as":           "answers",
			},
		}},
		{{
			Key: "$addFields",
			Value: bson.M{
				"answers": bson.M{
					"$sortArray": bson.M{
						"input":  "$answers",
						"sortBy": bson.M{"order_index": 1},
					},
				},
			},
		}},
		// 3. Lookup participant info
		{{
			Key: "$lookup",
			Value: bson.M{
				"from":         "event_participant",
				"localField":   "user_id",
				"foreignField": "user_id",
				"as":           "participant",
			},
		}},
		// 4. Unwind participant array
		{{Key: "$unwind", Value: "$participant"}},
		// 5. Exclude organizers
		{{Key: "$match", Value: bson.M{"participant.role": bson.M{"$ne": "organizer"}}}},
		// 6. Lookup user details
		{{
			Key: "$lookup",
			Value: bson.M{
				"from":         "users",
				"localField":   "user_id",
				"foreignField": "_id",
				"as":           "user",
			},
		}},
		// 7. Unwind user array
		{{Key: "$unwind", Value: "$user"}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("aggregation error: %w", err)
	}
	defer cursor.Close(ctx)

	var results []dto.AggregateResponse
	for cursor.Next(ctx) {
		var r dto.AggregateResponse
		if err := cursor.Decode(&r); err != nil {
			return nil, fmt.Errorf("failed to decode aggregation: %w", err)
		}
		results = append(results, r)
	}

	return results, nil
}

func CheckParticipantExists(ctx context.Context, eventID bson.ObjectID, userID bson.ObjectID) (bool, error) {
	count, err := database.DB.Collection("event_participant").CountDocuments(ctx, bson.M{"event_id": eventID, "user_id": userID})
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func FindFormByEventID(ctx context.Context, eventID string) (*models.Event_form, error) {
	objectID, err := bson.ObjectIDFromHex(eventID)
	if err != nil {
		return nil, fmt.Errorf("invalid form ID: %w", err)
	}

	var form models.Event_form
	err = database.DB.Collection("event_form").FindOne(ctx, bson.M{"event_id": objectID}).Decode(&form)
	if err != nil {
		return nil, err
	}

	return &form, nil
}