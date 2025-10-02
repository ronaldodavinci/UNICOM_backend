package repositories

import (
	"context"

	"github.com/pllus/main-fiber/tamarind/config"
	"github.com/pllus/main-fiber/tamarind/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type EventRepository struct {
	eventCol *mongo.Collection
	schedCol *mongo.Collection
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
	var docs []interface{}
	for _, sch := range s {
		docs = append(docs, sch)
	}
	_, err := r.schedCol.InsertMany(ctx, docs)
	return err
}

func (r *EventRepository) FindEvents(ctx context.Context, filter bson.M) ([]models.Event, error) {
	cur, err := r.eventCol.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var events []models.Event
	if err := cur.All(ctx, &events); err != nil {
		return nil, err
	}
	return events, nil
}