package repository

import (
	"context"

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

func (r *PolicyRepository) FindByPositionsAndAction(ctx context.Context, positions []string, action string) ([]models.Role, error) {
	cur, err := r.col.Find(ctx, bson.M{
		"enabled":      true,
		"position_key": bson.M{"$in": positions},
		"actions":      action,
	})
	if err != nil {
		return nil, err
	}
	var result []models.Role
	if err := cur.All(ctx, &result); err != nil {
		return nil, err
	}
	return result, nil
}
