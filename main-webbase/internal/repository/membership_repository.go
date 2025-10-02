package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"main-webbase/database"
	"main-webbase/internal/models"
)

type MembershipRepository struct {
	col *mongo.Collection
}

func NewMembershipRepository() *MembershipRepository {
	return &MembershipRepository{col: database.DB.Collection("memberships")}
}

func (r *MembershipRepository) FindByUser(ctx context.Context, userID any) ([]models.Membership, error) {
	filter := bson.M{"user_id": userID, "active": true}

	cur, err := r.col.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var memberships []models.Membership
	if err := cur.All(ctx, &memberships); err != nil {
		return nil, err
	}
	return memberships, nil
}

func (r *MembershipRepository) Insert(ctx context.Context, m models.Membership) error {
	_, err := r.col.InsertOne(ctx, m)
	return err
}

func (r *MembershipRepository) Delete(ctx context.Context, id any) error {
	_, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
