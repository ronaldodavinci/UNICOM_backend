package repositories

import (
	"context"
	"time"

	"github.com/pllus/main-fiber/tamarind/config"
	"github.com/pllus/main-fiber/tamarind/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MembershipRepository struct {
	col *mongo.Collection
}

func NewMembershipRepository() *MembershipRepository {
	return &MembershipRepository{
		col: config.DB.Collection("memberships"),
	}
}

func (r *MembershipRepository) Insert(ctx context.Context, m models.Membership) error {
	m.ID = primitive.NewObjectID()
	m.CreatedAt = time.Now()
	m.UpdatedAt = time.Now()
	_, err := r.col.InsertOne(ctx, m)
	return err
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

func (r *MembershipRepository) FindByUser(ctx context.Context, userID primitive.ObjectID) ([]models.Membership, error) {
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