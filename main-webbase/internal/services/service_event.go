package services

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"main-webbase/dto"
	"main-webbase/internal/models"
	repo "main-webbase/internal/repository"
)

// Use in CreateEventHandler
func CreateEventWithSchedules(body dto.EventRequestDTO, ctx context.Context) (models.Event, []models.EventSchedule, error) {
	now := time.Now().UTC()

	nodeID, err := bson.ObjectIDFromHex(body.NodeID)
	if err != nil {
		return models.Event{}, nil, fmt.Errorf("invalid NodeID: %w", err)
	}

	event := models.Event{
		ID:               bson.NewObjectID(),
		NodeID:           nodeID,
		Topic:            body.Topic,
		Description:      body.Description,
		MaxParticipation: body.MaxParticipation,
		PostedAs:         body.PostedAs,
		Visibility:       body.Visibility,
		OrgOfContent:     body.OrgOfContent,
		Status:           body.Status,
		Have_form: 		  false,
		CreatedAt:        &now,
		UpdatedAt:        &now,
	}

	// Insert event
	if err := repo.InsertEvent(ctx, event); err != nil {
		return models.Event{}, nil, err
	}

	// Prepare schedules
	var schedules []models.EventSchedule
	for _, s := range body.Schedules {
		schedules = append(schedules, models.EventSchedule{
			ID:          bson.NewObjectID(),
			EventID:     event.ID,
			Date:        s.Date,
			Time_start:  s.Time_start,
			Time_end:    s.Time_end,
			Location:    &s.Location,
			Description: &s.Description,
		})
	}

	// Insert schedules
	if err := repo.InsertSchedules(ctx, schedules); err != nil {
		return event, nil, err
	}
	return event, schedules, nil
}

// Use in GetAllVisibleEventHandler
func GetVisibleEvents(viewerID bson.ObjectID, ctx context.Context, orgSets []string) ([]map[string]interface{}, error) {
	// Get all event
	events, err := repo.GetEvent(ctx)
	if err != nil {
		return nil, err
	}

	// Get ID of Visible Event
	var eventIDlist []bson.ObjectID
	for _, ev := range events {
		if CheckVisibleEvent(&ev, orgSets) {
			eventIDlist = append(eventIDlist, ev.ID)
		}
	}

	// Fetch schedule of visible event
	schedules, err := repo.GetSchedulesByEvent(ctx, eventIDlist)
	if err != nil {
		return nil, err
	}

	//Group Then Send
	schedMap := make(map[bson.ObjectID][]models.EventSchedule)
	for _, s := range schedules {
		schedMap[s.EventID] = append(schedMap[s.EventID], s)
	}

	var result []map[string]interface{}
	for _, ev := range events {
		if !CheckVisibleEvent(&ev, orgSets) {
			continue
		}
		result = append(result, map[string]interface{}{
			"event":     ev,
			"schedules": schedMap[ev.ID],
		})
	}

	return result, nil
}

func CheckVisibleEvent(event *models.Event, userOrgs []string) bool {
	if event.Status == "hidden" {
		return false
	}

	v := event.Visibility
	switch v.Access {
	// 1. Access = public
	case "public":
		return true

	// 2. Access = Selected path only
	case "org":
		if len(v.Audience) == 0 {
			return false
		}

		subtreeSet := map[string]struct{}{}
		for _, s := range userOrgs {
			subtreeSet[s] = struct{}{}
		}

		for _, a := range v.Audience {
			if _, ok := subtreeSet[a.OrgPath]; ok {
				return true
			}
		}
		return false
	}
	return false
}
