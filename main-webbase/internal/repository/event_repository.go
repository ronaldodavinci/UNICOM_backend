package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"main-webbase/database"
	"main-webbase/internal/models"
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
	cursor, err := database.DB.Collection("events").Find(ctx, bson.M{"status": bson.M{"$ne": "hidden"}})
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
