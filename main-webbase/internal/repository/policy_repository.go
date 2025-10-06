package repository

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"main-webbase/database"
	"main-webbase/internal/models"
)

type PolicyRepository struct {
	col *mongo.Collection
}

func NewPolicyRepository() *PolicyRepository {
	return &PolicyRepository{col: database.DB.Collection("policies")}
}

func (r *PolicyRepository) Insert(ctx context.Context, p models.Policy) error {
	_, err := r.col.InsertOne(ctx, p)
	return err
}

func (r *PolicyRepository) FindAll(ctx context.Context) ([]models.Policy, error) {
	cur, err := r.col.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var policies []models.Policy
	if err := cur.All(ctx, &policies); err != nil {
		return nil, err
	}
	return policies, nil
}


func (r *PolicyRepository) FindByPositionsAndAction(ctx context.Context, posKeys []string, action string) ([]models.Policy, error) {
	filter := bson.M{
		"position_key": bson.M{"$in": posKeys},
		"actions":      action,
		"enabled":      true,
	}

	cur, err := r.col.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var policies []models.Policy
	if err := cur.All(ctx, &policies); err != nil {
		return nil, err
	}
	return policies, nil
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