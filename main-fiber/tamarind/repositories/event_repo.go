package repositories

import (
	"context"

	"github.com/pllus/main-fiber/tamarind/config/config"
	"github.com/pllus/main-fiber/tamarind/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type EventRepository struct {
	eventCol   *mongo.Collection
	schedCol   *mongo.Collection
}

func NewEventRepository() *EventRepository {
	return &EventRepository{
		eventCol: config.DB.Collection("events"),
		schedCol: config.DB.Collection("event_schedules"),
	}
}

func (r *EventRepository) InsertEvent(ctx context.Context, e models.Event) error {
	_, err := r.eventCol.InsertOne(ctx, e)
	return err
}

func (r *EventRepository) InsertSchedules(ctx context.Context, s []models.EventSchedule) error {
	if len(s) == 0 {
		return nil
	}
	_, err := r.schedCol.InsertMany(ctx, s)
	return err
}

func (r *EventRepository) FindEvents(ctx context.Context, filter bson.M) ([]models.Event, error) {
	cur, err := r.eventCol.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	var events []models.Event
	if err := cur.All(ctx, &events); err != nil {
		return nil, err
	}
	return events, nil
}

func (r *EventRepository) FindSchedulesByEvent(ctx context.Context, ids []any) ([]models.EventSchedule, error) {
	cur, err := r.schedCol.Find(ctx, bson.M{"event_id": bson.M{"$in": ids}})
	if err != nil {
		return nil, err
	}
	var schedules []models.EventSchedule
	if err := cur.All(ctx, &schedules); err != nil {
		return nil, err
	}
	return schedules, nil
}
