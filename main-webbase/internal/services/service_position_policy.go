package services

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"main-webbase/dto"
	"main-webbase/internal/models"
	"main-webbase/database"
	repo "main-webbase/internal/repository"
)

func CreatePositionWithPolicy(body dto.PositionCreateDTO, ctx context.Context) (*models.Position, *models.Policy, error) {
	now := time.Now().UTC()

	if body.Key == "" {
		return nil, nil, errors.New("position key cannot be empty")
	}
	if body.Scope == nil || body.Scope.OrgPath == "" {
		return nil, nil, errors.New("scope org_path cannot be empty")
	}

	position := &models.Position{
		ID:          bson.NewObjectID(),
		Key:         body.Key,
		Display:     body.Display,
		Rank:        body.Rank,
		Status:      body.Status,
		Constraints: *body.Constraints,
		Scope:       *body.Scope,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	positionCol := database.DB.Collection("positions")
	if _, err := positionCol.InsertOne(ctx, position); err != nil {
		return nil, nil, err
	}

	policy := &models.Policy{
		ID:          bson.NewObjectID(),
		PositionKey: position.Key,
		Scope:       body.Policy.Type, // “exact” or “subtree”
		OrgPrefix:   body.Scope.OrgPath,
		Actions:     body.Policy.Actions,
		Enabled:     body.Policy.Enabled,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	policyCol := database.DB.Collection("policies")
	if _, err := policyCol.InsertOne(ctx, policy); err != nil {
		// Rollback if error
		_, _ = positionCol.DeleteOne(ctx, bson.M{"_id": position.ID})
		return nil, nil, err
	}
	
	return position, policy, nil
}

func FindPolicyByKeyandPath(ctx context.Context, key string, path string) (*models.Policy, error) {
	col := database.DB.Collection("policies")
	filter := bson.M{
		"position_key": key,
		"org_prefix":   path,
	}
	var policy models.Policy
	err := col.FindOne(ctx, filter).Decode(&policy)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}

	return &policy, nil
}

func MyUserPolicy(ctx context.Context, uid string) ([]models.Policy, error) {
	memberships, err := repo.GetUserMemberships(ctx, uid)
	if err != nil {
		return nil, err
	}

	if len(memberships) == 0 {
		return []models.Policy{}, nil
	}

	var orFilters []bson.M
	for _, m := range memberships {
		orFilters = append(orFilters, bson.M{
			"position_key": m.PositionKey,
			"org_prefix":   m.OrgPath,
		})
	}

	policyCol := database.DB.Collection("policies")
	cursor, err := policyCol.Find(ctx, bson.M{"$or": orFilters})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var policies []models.Policy
	if err := cursor.All(ctx, &policies); err != nil {
		return nil, err
	}

	return policies, nil
}

func UpdatedPolicy(ctx context.Context, policy *models.Policy) (error) {
	col := database.DB.Collection("policies")

	_, err := col.UpdateOne(ctx, 
		bson.M{"_id": policy.ID},
		bson.M{"$set": bson.M{
			"actions":    policy.Actions,
			"updated_at": time.Now().UTC(),
			"enabled":    policy.Enabled,
		}},
	)
	return err
}