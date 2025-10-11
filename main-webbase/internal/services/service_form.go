package services

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"main-webbase/database"
	"main-webbase/dto"
	"main-webbase/internal/models"
	repo "main-webbase/internal/repository"
)

func InitializeFormService(eventID string, ctx context.Context) (models.Event_form, error) {
	now := time.Now().UTC()

	EventID, err := bson.ObjectIDFromHex(eventID)
	if err != nil {
		return models.Event_form{}, fmt.Errorf("invalid EventID: %w", err)
	}

	var exist models.Event_form
	err = database.DB.Collection("event_form").FindOne(ctx, bson.M{"event_id": EventID}).Decode(&exist)
	if err == nil {
		if err := repo.UpdateEvent(ctx, EventID, bson.M{"have_form": true}); err != nil {
			return exist, err
		}
	}

	form := models.Event_form{
		ID:        bson.NewObjectID(),
		Event_ID:  EventID,
		CreatedAt: &now,
		UpdatedAt: &now,
	}

	if err := repo.InitializeForm(ctx, form); err != nil {
		return models.Event_form{}, err
	}

	if err := repo.UpdateEvent(ctx, EventID, bson.M{"have_form": true}); err != nil {
		return models.Event_form{}, err
	}

	return form, nil
}

func DisableFormService(eventID string, ctx context.Context) error {
	EventID, err := bson.ObjectIDFromHex(eventID)
	if err != nil {
		return fmt.Errorf("invalid EventID: %w", err)
	}

	if err := repo.UpdateEvent(ctx, EventID, bson.M{"have_form": false}); err != nil {
		return err
	}

	return nil
}

func CreateFormQuestion(body dto.FormQuestionCreateDTO, ctx context.Context) ([]models.Event_form_question, error) {
	now := time.Now().UTC()

	formID, err := bson.ObjectIDFromHex(body.FormID)
	if err != nil {
		return nil, fmt.Errorf("invalid formID: %w", err)
	}

	// Delete all previous form question
	if err := repo.DeleteQuestionsByFormID(ctx, formID); err != nil {
		return nil, fmt.Errorf("failed to delete existing questions: %w", err)
	}

	// Creating new set of question
	var newQuestions []models.Event_form_question
	for _, q := range body.Questions {
		newQuestions = append(newQuestions, models.Event_form_question{
			ID:            bson.NewObjectID(),
			Form_ID:       formID,
			Question_text: q.QuestionText,
			Required:      q.Required,
			OrderIndex:    q.OrderIndex,
			CreatedAt:     &now,
		})
	}

	if err := repo.InsertFormQuestions(ctx, newQuestions); err != nil {
		return nil, fmt.Errorf("failed to insert new questions: %w", err)
	}

	return newQuestions, nil
}

func GetFormQuestion(ctx context.Context, formID string) ([]models.Event_form_question, error) {
	FormID, err := bson.ObjectIDFromHex(formID)
	if err != nil {
		return nil, fmt.Errorf("invalid formID: %w", err)
	}

	questions, err := repo.FindQuestionsByFormID(ctx, FormID)
	if err != nil {
		return nil, fmt.Errorf("failed to get questions: %w", err)
	}

	return questions, nil
}

func SubmitFormResponse(body dto.FormResponseSubmitDTO, ctx context.Context, userID string) ([]models.Event_form_answer, error) {
	formID, err := bson.ObjectIDFromHex(body.FormID)
	if err != nil {
		return nil, fmt.Errorf("invalid formID: %w", err)
	}

	UserID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid userID: %w", err)
	}

	now := time.Now().UTC()
	response := models.Event_response{
		ID:       bson.NewObjectID(),
		Form_ID:  formID,
		User_ID:  UserID,
		SubmitAt: &now,
	}

	if err := repo.SubmitResponse(ctx, response); err != nil {
		return nil, fmt.Errorf("failed to Create form response: %w", err)
	}

	// For insert many
	answers := make([]models.Event_form_answer, 0, len(body.Answers))
	docs := make([]interface{}, 0, len(body.Answers))

	for _, ans := range body.Answers {
		qid, err := bson.ObjectIDFromHex(ans.QuestionID)
		if err != nil {
			return nil, fmt.Errorf("invalid QuestionID: %w", err)
		}

		answer := models.Event_form_answer{
			ID:           bson.NewObjectID(),
			Question_ID:  qid,
			Response_ID:  response.ID,
			Answer_value: ans.AnswerValue,
			CreatedAt:    &now,
		}

		answers = append(answers, answer)
		docs = append(docs, answer)
	}

	if err := repo.InsertAnswers(ctx, docs); err != nil {
		return nil, fmt.Errorf("failed to insert answers: %w", err)
	}

	form, err := repo.FindFormByID(ctx, body.FormID)
	if err != nil {
		return nil, fmt.Errorf("failed to find form: %w", err)
	}

	participant := models.Event_participant{
		ID:          bson.NewObjectID(),
		Event_ID:    form.Event_ID,
		User_ID:     UserID,
		Response_ID: response.ID,
		Status:      "Stall",
		Role:        "Participant",
		CreatedAt:   &now,
	}

	if err := repo.AddParticipant(ctx, participant); err != nil {
		return nil, fmt.Errorf("failed to add participant: %w", err)
	}

	return answers, nil
}

func GetAllResponse(ctx context.Context, formID string) (dto.FormMatrixResponseDTO, error) {
	var result dto.FormMatrixResponseDTO

	FormID, err := bson.ObjectIDFromHex(formID)
	if err != nil {
		return result, fmt.Errorf("invalid formID: %w", err)
	}

	questions, err := repo.FindQuestionsByFormID(ctx, FormID)
	if err != nil {
		return result, fmt.Errorf("failed to get questions: %w", err)
	}

	questionDTOs := make([]dto.QuestionDTO, len(questions))
	questionIndexMap := make(map[string]int)
	for i, q := range questions {
		questionDTOs[i] = dto.QuestionDTO{
			ID:   q.ID.Hex(),
			Text: q.Question_text,
		}
		questionIndexMap[q.ID.Hex()] = i
	}
	result.Questions = questionDTOs

	aggResponseAnswer, err := repo.AggregateUserResponse(ctx, FormID)
	if err != nil {
		return result, fmt.Errorf("failed to get Answer List: %w", err)
	}

	for _, r := range aggResponseAnswer {
		user := dto.UserAnswersDTO{
			UserID:    r.User.ID.Hex(),
			FirstName: r.User.FirstName,
			LastName:  r.User.LastName,
			Status:    r.Participant.Status,
			Answers:   make([]string, len(questions)),
		}

		for _, ans := range r.Answers {
			if idx, ok := questionIndexMap[ans.Question_ID.Hex()]; ok {
				user.Answers[idx] = ans.Answer_value
			}
		}

		result.Responses = append(result.Responses, user)
	}

	return result, nil
}

func UpdateParticipantStatus(ctx context.Context, body dto.UpdateParticipantStatusDTO) error {
	col := database.DB.Collection("event_participant")

	_, err := col.UpdateOne(ctx,
		bson.M{"user_id": body.UserID,
			"event_id": body.EventID},
		bson.M{"$set": bson.M{
			"status": body.Status,
		}},
	)

	return err
}

func GetParticipantStatus(ctx context.Context, userID string, eventID string) (*models.Event_participant, error) {
	col := database.DB.Collection("event_participant")

	EventID, err := bson.ObjectIDFromHex(eventID)
	if err != nil {
		return nil, err
	}

	UserID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}
	filter := bson.M{"user_id": UserID, "event_id": EventID}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var userStatus models.Event_participant
	err = col.FindOne(ctx, filter).Decode(&userStatus)
	if err != nil {
		return nil, err
	}

	return &userStatus, nil
}
