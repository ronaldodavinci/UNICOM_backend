package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"main-webbase/database"
	"main-webbase/internal/models"
)

type OrgUnitRepository struct {
	col *mongo.Collection
}

func NewOrgUnitRepository() *OrgUnitRepository {
	return &OrgUnitRepository{col: database.DB.Collection("org_units")}
}

func FindByOrgPath(ctx context.Context, path string) (*models.OrgUnitNode, error) {
	col := database.DB.Collection("org_units")

	var node models.OrgUnitNode
	err := col.FindOne(ctx, bson.M{"org_path": path}).Decode(&node)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	
	return &node, nil
}

func NodeCreate(ctx context.Context, node models.OrgUnitNode) error {
	_, err := database.DB.Collection("org_units").InsertOne(ctx, node)
	return err
}

func FindByPrefix(ctx context.Context, path string) ([]models.OrgUnitNode, error) {
	col := database.DB.Collection("org_units")

	filter := bson.M{
		"$or": []bson.M{
			{"ancestors": path},
			{"org_path": path},
		},
	}

	cur, err := col.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var results []models.OrgUnitNode
	if err := cur.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

func GetOrgByID(ctx context.Context, OrgID bson.ObjectID) (*models.OrgUnitNode, error) {
	collection := database.DB.Collection("org_units")
	var orgnode models.OrgUnitNode

	err := collection.FindOne(ctx, bson.M{"_id": OrgID}).Decode(&orgnode)
	if err != nil {
		return nil, err
	}

	return &orgnode, nil
}