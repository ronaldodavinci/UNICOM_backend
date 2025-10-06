package repository

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"main-webbase/database"
	"main-webbase/internal/models"
)

type PositionRepository struct {
	col *mongo.Collection
}

func NewPositionRepository() *PositionRepository {
	return &PositionRepository{
		col: database.DB.Collection("positions"),
	}
}

func FindPositionByKeyandPath(ctx context.Context, key string, path string) (*models.Position, error) {
	col := database.DB.Collection("positions")

	var position models.Position
	err := col.FindOne(ctx, bson.M{"key": key, "scope.org_path": path,}).Decode(&position)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("position not found (key=%s, org_path=%s)", key, path)
		}
		return nil, fmt.Errorf("error finding position: %w", err)
	}

	return &position, nil
}

func (r *PositionRepository) FindAll(ctx context.Context) ([]models.Position, error) {
	cur, err := r.col.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var positions []models.Position
	if err := cur.All(ctx, &positions); err != nil {
		return nil, err
	}
	return positions, nil
}
