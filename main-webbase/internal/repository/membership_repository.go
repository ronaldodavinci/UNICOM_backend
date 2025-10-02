package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
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

func (r *MembershipRepository) FindByUser(ctx context.Context, userID bson.ObjectID) ([]models.Membership, error) {
	cur, err := r.col.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var memberships []models.Membership
	if err := cur.All(ctx, &memberships); err != nil {
		return nil, err
	}
	return memberships, nil
}

func (r *MembershipRepository) FindAll(ctx context.Context) ([]models.Membership, error) {
	cur, err := r.col.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var memberships []models.Membership
	if err := cur.All(ctx, &memberships); err != nil {
		return nil, err
	}
	return memberships, nil
}

func (r *MembershipRepository) Insert(ctx context.Context, m models.Membership) error {
	m.ID = bson.NewObjectID()
	m.CreatedAt = time.Now()
	m.UpdatedAt = time.Now()
	_, err := r.col.InsertOne(ctx, m)
	return err
}

func (r *MembershipRepository) Delete(ctx context.Context, id any) error {
	_, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
