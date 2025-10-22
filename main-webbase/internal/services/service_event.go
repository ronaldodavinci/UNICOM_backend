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
func CreateEventWithSchedules(body dto.EventRequestDTO, ctx context.Context) (dto.EventCreateResult, error) {
	now := time.Now().UTC()

	nodeID, err := bson.ObjectIDFromHex(body.NodeID)
	if err != nil {
		return dto.EventCreateResult{}, fmt.Errorf("invalid NodeID: %w", err)
	}

	event := models.Event{
		ID:               bson.NewObjectID(),
		NodeID:           nodeID,
		Topic:            body.Topic,
		Description:      body.Description,
		PictureURL:       &body.PictureURL,
		MaxParticipation: body.MaxParticipation,
		PostedAs:         body.PostedAs,
		Visibility:       body.Visibility,
		OrgOfContent:     body.OrgOfContent,
		Status:           body.Status,
		Have_form:        body.Have_form,
		CreatedAt:        &now,
		UpdatedAt:        &now,
	}

	// Insert event
	if err := repo.InsertEvent(ctx, event); err != nil {
		return dto.EventCreateResult{}, fmt.Errorf("failed to insert event: %w", err)
	}

	// Prepare schedules
	var schedules []models.EventSchedule
	for _, s := range body.Schedules {
		schedules = append(schedules, models.EventSchedule{
			ID:          bson.NewObjectID(),
			EventID:     event.ID,
			Date:        s.Date,
			Time_start:  s.TimeStart,
			Time_end:    s.TimeEnd,
			Location:    &s.Location,
			Description: &s.Description,
		})
	}

	if len(schedules) > 0 {
		if err := repo.InsertSchedules(ctx, schedules); err != nil {
			return dto.EventCreateResult{}, fmt.Errorf("failed to insert schedules: %w", err)
		}
	}

	var formID string
	if body.Have_form {
		form := models.Event_form{
			ID:        bson.NewObjectID(),
			Event_ID:  event.ID,
			CreatedAt: &now,
			UpdatedAt: &now,
		}

		if err := repo.InitializeForm(ctx, form); err != nil {
			return dto.EventCreateResult{}, fmt.Errorf("failed to initialize form: %w", err)
		}
		formID = form.ID.Hex()
	}

	members, err := repo.FindMembershipByManageEvent(ctx, body.OrgOfContent)
	if err != nil {
		return dto.EventCreateResult{}, fmt.Errorf("failed to find memberships: %w", err)
	}

	var participants []models.Event_participant
	for _, member := range members {
		participants = append(participants, models.Event_participant{
			ID:        bson.NewObjectID(),
			Event_ID:  event.ID,
			User_ID:   member.UserID,
			Status:    "accept",
			Role:      "organizer",
			CreatedAt: &now,
		})
	}
	if len(participants) > 0 {
		if err := repo.AddListParticipant(ctx, participants); err != nil {
			return dto.EventCreateResult{}, fmt.Errorf("failed to add initial participants: %w", err)
		}
	}

	return dto.EventCreateResult{
		Event:        event,
		Schedules:    schedules,
		FormID:       formID,
		OrganizerCnt: len(participants),
	}, nil
}

// Use in GetAllVisibleEventHandler
func GetVisibleEvents(viewerID bson.ObjectID, ctx context.Context, orgSets []string) ([]dto.EventFeed, error) {
	// Get all event
	events, err := repo.GetEvent(ctx)
	if err != nil {
		return nil, err
	}

	// Get ID of Visible Event
	var visibleEvents []models.Event
	var eventIDlist []bson.ObjectID
	for _, ev := range events {
		if CheckVisibleEvent(ctx, &ev, orgSets, viewerID) {
			visibleEvents = append(visibleEvents, ev)
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

	// Get every event participant count
	participantCounts, err := repo.GetAllEventParticipant(ctx, eventIDlist)
	if err != nil {
		participantCounts = make(map[bson.ObjectID]int)
	}

	var result []dto.EventFeed
	for _, ev := range visibleEvents {

		eventDetail := dto.EventFeed{
			EventID:              ev.ID.Hex(),
			OrgPath:              ev.OrgOfContent,
			Topic:                ev.Topic,
			Description:          ev.Description,
			PictureURL:           ev.PictureURL,
			MaxParticipation:     ev.MaxParticipation,
			CurrentParticipation: participantCounts[ev.ID],
			PostedAs:             ev.PostedAs,
			Visibility:           ev.Visibility,
			Status:               ev.Status,
			Have_form:            ev.Have_form,
			Schedules:            schedMap[ev.ID],
		}

		result = append(result, eventDetail)
	}

	return result, nil
}

func CheckVisibleEvent(ctx context.Context, event *models.Event, userOrgs []string, userID bson.ObjectID) bool {
	if event.Status == "inactive" {
		return false
	}

	if event.Status == "active" {
		v := event.Visibility
		if v == nil {
			return false
		}
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
	}

	if event.Status == "draft" {
		subtreeSet := map[string]struct{}{} 
		for _, s := range userOrgs { 
			subtreeSet[s] = struct{}{} 
		} 
		if _, ok := subtreeSet[event.OrgOfContent]; ok { 
			userStatus, err := GetParticipantStatus(ctx, userID.Hex(), event.ID.Hex())
			if err != nil {
				return false
			}
			if userStatus.Role == "organizer" {
				return true
			}
		}
		return false
	}

	return false
}

func GetEventDetail(eventID string, ctx context.Context) (dto.EventDetail, error) {
	// Convert string ID to BSON ObjectID
	EventID, err := bson.ObjectIDFromHex(eventID)
	if err != nil {
		return dto.EventDetail{}, fmt.Errorf("invalid EventID: %w", err)
	}

	// Fetch event
	event, err := repo.GetEventByID(ctx, EventID)
	if err != nil {
		return dto.EventDetail{}, fmt.Errorf("failed to get event: %w", err)
	}

	// Fetch schedules
	schedules, err := repo.GetEventScheduleByID(ctx, EventID)
	if err != nil {
		return dto.EventDetail{}, fmt.Errorf("failed to get schedules: %w", err)
	}

	// Fetch current participant count
	current_participant, err := repo.GetTotalParticipant(ctx, EventID)
	if err != nil {
		return dto.EventDetail{}, fmt.Errorf("failed to get participants: %w", err)
	}

	var formMatrix dto.FormMatrixResponseDTO

	if event.Have_form {
		// Fetch all form responses
		formMatrix, err = GetAllResponse(ctx, eventID)
		if err != nil {
			return dto.EventDetail{}, fmt.Errorf("failed to get form responses: %w", err)
		}
	}

	eventDetail := dto.EventDetail{
		EventID:              event.ID.Hex(),
		OrgPath:              event.OrgOfContent,
		Topic:                event.Topic,
		Description:          event.Description,
		PictureURL:           event.PictureURL,
		MaxParticipation:     event.MaxParticipation,
		CurrentParticipation: current_participant,
		PostedAs:             event.PostedAs,
		Visibility:           event.Visibility,
		Status:               event.Status,
		Have_form:            event.Have_form,
		Schedules:            schedules,
		FormMatrixResponse:   formMatrix,
	}

	return eventDetail, nil
}


func ParticipateEvent(eventID string, uid string, ctx context.Context) error {
	now := time.Now().UTC()

	eventObjID, err := bson.ObjectIDFromHex(eventID)
	if err != nil {
		return fmt.Errorf("invalid EventID: %w", err)
	}

	userObjID, err := bson.ObjectIDFromHex(uid)
	if err != nil {
		return fmt.Errorf("invalid UserID: %w", err)
	}

	event, err := repo.GetEventByID(ctx, eventObjID)
	if err != nil {
		return fmt.Errorf("failed to find event: %w", err)
	}
	if event.Have_form {
		return fmt.Errorf("this event requires form submission to join")
	}

	exists, err := repo.CheckParticipantExists(ctx, eventObjID, userObjID)
	if err != nil {
		return fmt.Errorf("failed to check participant: %w", err)
	}
	if exists {
		return fmt.Errorf("user already joined this event")
	}

	participant := models.Event_participant{
		ID:        bson.NewObjectID(),
		Event_ID:  eventObjID,
		User_ID:   userObjID,
		Status:    "accept",
		Role:      "participant",
		CreatedAt: &now,
	}

	if err := repo.AddParticipant(ctx, participant); err != nil {
		return fmt.Errorf("failed to add participant: %w", err)
	}

	return nil
}
