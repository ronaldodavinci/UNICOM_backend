package repository

import (
	"context"
	"time"
	"strings"
	"regexp"

	"main-webbase/database"
	"main-webbase/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// internal/repository/event_repo.go

type EventFilter struct {
    Roles    []string
    Q        string
}

func GetEventsFilter(ctx context.Context, f EventFilter) ([]models.Event, error) {
    coll := database.DB.Collection("events")

    and := []bson.M{
        {"status": bson.M{"$ne": "inactive"}},
    }

	conds := make([]bson.M, 0, len(f.Roles)*2)

	if len(f.Roles) > 0 {
		for _, r := range f.Roles {
			r = strings.TrimSpace(r)
			if r == "" {
				continue
			}

			if strings.HasPrefix(r, "/") {
				// เป็น org path หรือ prefix
				if strings.HasSuffix(r, "/*") {
					// subtree prefix: ^\Qprefix\E(/|$)
					prefix := strings.TrimSuffix(r, "/*")
					re := "^" + regexp.QuoteMeta(prefix)
					conds = append(conds, bson.M{
						"org_of_content": bson.M{"$regex": re},
					})
				} else {
					// exact path
					conds = append(conds, bson.M{
						"org_of_content": r,
					})
				}
			} else {
				// เป็นตำแหน่ง (postedas.position_key) — match แบบเต็มคำ ไม่สนตัวพิมพ์เล็กใหญ่
				re := "^" + regexp.QuoteMeta(r) + "$"
				conds = append(conds, bson.M{
					"postedas.position_key": bson.M{"$regex": re, "$options": "i"},
				})
			}
		}

		if len(conds) == 1 {
			and = append(and, conds[0])
		} else if len(conds) > 1 {
			and = append(and, bson.M{"$or": conds})
		}
	}

    if f.Q != "" {
        rx := bson.M{"$regex": f.Q, "$options": "i"}
        and = append(and, bson.M{
            "$or": []bson.M{
                {"topic":       rx},
                {"description": rx},
            },
        })
    }

    filter := bson.M{}
    if len(and) > 0 {
        filter = bson.M{"$and": and}
    }

    cur, err := coll.Find(ctx, filter)
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

	if len(eventIDlist) == 0 {
		return []models.EventSchedule{}, nil
	}

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