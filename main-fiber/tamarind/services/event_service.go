package services

import (
	"context"
	"time"

	"github.com/pllus/main-fiber/models"
	"github.com/pllus/main-fiber/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EventService struct {
	eventRepo      *repositories.EventRepository
	membershipRepo *repositories.MembershipRepository
}

func NewEventService(e *repositories.EventRepository, m *repositories.MembershipRepository) *EventService {
	return &EventService{eventRepo: e, membershipRepo: m}
}

// CreateEventWithSchedules handles event + schedule creation
func (s *EventService) CreateEventWithSchedules(ctx context.Context, req models.Event, schedules []models.EventSchedule) error {
	req.ID = primitive.NewObjectID()
	now := time.Now().UTC()
	req.CreatedAt = &now
	req.UpdatedAt = &now

	if err := s.eventRepo.InsertEvent(ctx, req); err != nil {
		return err
	}
	return s.eventRepo.InsertSchedules(ctx, schedules)
}

// GetVisibleEvents fetches events the user is allowed to see
func (s *EventService) GetVisibleEvents(ctx context.Context, viewerID primitive.ObjectID, orgs []string) ([]models.Event, error) {
	events, err := s.eventRepo.FindEvents(ctx, map[string]any{"status": map[string]any{"$ne": "hidden"}})
	if err != nil {
		return nil, err
	}

	// naive filter: public or org match
	var visible []models.Event
	for _, ev := range events {
		if ev.Status == "public" || containsOrg(orgs, ev.OrgOfContent) {
			visible = append(visible, ev)
		}
	}
	return visible, nil
}

func containsOrg(orgs []string, path string) bool {
	for _, o := range orgs {
		if o == path {
			return true
		}
	}
	return false
}
