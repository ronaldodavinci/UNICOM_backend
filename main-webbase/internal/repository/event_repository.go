package repository

import (
	"context"
	"time"

	"main-webbase/database"
	"main-webbase/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Use in CreateEventWithSchedules
func InsertEvent(ctx context.Context, event models.Event) error {
	_, err := database.DB.Collection("events").InsertOne(ctx, event)
	return err
}

func InsertSchedules(ctx context.Context, schedules []models.EventSchedule) error {
	if len(schedules) == 0 {
		return nil
	}
	_, err := database.DB.Collection("event_schedules").InsertMany(ctx, schedules)
	return err
}

// Use in GetVisibleEvents
func GetEvent(ctx context.Context) ([]models.Event, error) {
	cursor, err := database.DB.Collection("events").Find(ctx, bson.M{"status": bson.M{"$ne": "inactive"}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []models.Event
	if err := cursor.All(ctx, &events); err != nil {
		return nil, err
	}
	return events, nil
}

func GetSchedulesByEvent(ctx context.Context, eventIDlist []bson.ObjectID) ([]models.EventSchedule, error) {
	collection := database.DB.Collection("event_schedules")

	filter := bson.M{"event_id": bson.M{"$in": eventIDlist}}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var schedules []models.EventSchedule
	if err := cursor.All(ctx, &schedules); err != nil {
		return nil, err
	}

	return schedules, nil
}

// Get Event Detail by EventID
func GetEventByID(ctx context.Context, EventID bson.ObjectID) (*models.Event, error) {
	collection := database.DB.Collection("events")
	var event models.Event

	err := collection.FindOne(ctx, bson.M{"_id": EventID}).Decode(&event)
	if err != nil {
		return nil, err
	}

	return &event, nil
}

func GetEventScheduleByID(ctx context.Context, EventID bson.ObjectID) ([]models.EventSchedule, error) {
	collection := database.DB.Collection("event_schedules")

	cursor, err := collection.Find(ctx, bson.M{"event_id": EventID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var schedules []models.EventSchedule
	if err := cursor.All(ctx, &schedules); err != nil {
		return nil, err
	}
	return schedules, nil
}

func FindEventForm(ctx context.Context, EventID bson.ObjectID) (*models.Event_form, error) {
	collection := database.DB.Collection("event_form")
	var event_form models.Event_form

	err := collection.FindOne(ctx, bson.M{"event_id": EventID}).Decode(&event_form)
	if err != nil {
		return nil, err
	}

	return &event_form, nil
}

func GetTotalParticipant(ctx context.Context, eventID bson.ObjectID) (int, error) {
	count, err := database.DB.Collection("event_participant").CountDocuments(ctx, bson.M{"event_id": eventID, "status": "accept", "role": "participant"})
	if err != nil {
		return 0, err
	}

	return int(count), nil
}


func FindAcceptedParticipants(ctx context.Context, eventID bson.ObjectID) ([]bson.ObjectID, error) {
	collection := database.DB.Collection("event_form_response")

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"event_id": eventID,
			"status":   "accept",
		}}},
		{{Key: "$project", Value: bson.M{
			"participant_id": 1,
			"_id":            0,
		}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var participantIDs []bson.ObjectID
	for cursor.Next(ctx) {
		var doc struct {
			ParticipantID bson.ObjectID `bson:"participant_id"`
		}
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		participantIDs = append(participantIDs, doc.ParticipantID)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return participantIDs, nil
}


func DeleteEventScheduleByID(ctx context.Context, eventID bson.ObjectID) error {
	collection := database.DB.Collection("event_schedules")
	_, err := collection.DeleteMany(ctx, bson.M{"event_id": eventID})
	return err
}

func UpdateEvent(ctx context.Context, eventID bson.ObjectID, updates bson.M) error {
	collection := database.DB.Collection("events")
	updates["updated_at"] = time.Now().UTC()

	_, err := collection.UpdateOne(ctx, bson.M{"_id": eventID}, bson.M{"$set": updates})
	return err
}