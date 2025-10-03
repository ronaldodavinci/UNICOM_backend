package repositories

import (
	"context"

	"github.com/pllus/main-fiber/tamarind/config"
	"github.com/pllus/main-fiber/tamarind/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type PolicyRepository struct {
	col *mongo.Collection
}

func NewPolicyRepository() *PolicyRepository {
	return &PolicyRepository{
		col: config.DB.Collection("policies"),
	}
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