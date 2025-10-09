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

func InitializeFormService(body dto.FormCreateDTO, ctx context.Context) (models.Event_form, error) {
	now := time.Now().UTC()

	eventID, err := bson.ObjectIDFromHex(body.EventID)
	if err != nil {
		return models.Event_form{}, fmt.Errorf("invalid EventID: %w", err)
	}

	form := models.Event_form{
		ID:			bson.NewObjectID(),
		Event_ID:	eventID,
		OrgPath: 	body.OrgPath,
		Status:		"Draft",
		CreatedAt:	&now,
		UpdatedAt: 	&now,
	}

	if err := repo.InitializeForm(ctx, form); err != nil {
		return models.Event_form{}, err
	}

	return form, nil
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
			ID:				bson.NewObjectID(),
			Form_ID:		formID,
			Question_text:	q.QuestionText,
			Required:		q.Required,
			OrderIndex:		q.OrderIndex,
			CreatedAt:		&now,
		})
	}

	if err := repo.InsertFormQuestions(ctx, newQuestions); err != nil {
		return nil, fmt.Errorf("failed to insert new questions: %w", err)
	}

	return newQuestions, nil
}