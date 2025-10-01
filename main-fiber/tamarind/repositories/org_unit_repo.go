package repositories

import (
	"context"

	"github.com/pllus/main-fiber/tamarind/config"
	"github.com/pllus/main-fiber/tamarind/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type OrgUnitRepository struct {
	col *mongo.Collection
}

func NewOrgUnitRepository() *OrgUnitRepository {
	return &OrgUnitRepository{col: config.DB.Collection("org_units")}
}

func (r *OrgUnitRepository) Find(ctx context.Context, filter bson.M) ([]models.OrgUnit, error) {
	cur, err := r.col.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	var result []models.OrgUnit
	if err := cur.All(ctx, &result); err != nil {
		return nil, err
	}
	return result, nil
}
