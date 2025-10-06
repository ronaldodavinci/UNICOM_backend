package repository

import (
	"context"
	"time"
	"fmt"

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

func InsertMembership(ctx context.Context, m models.Membership) error {
	// 1. Check if OrgPath exists
	orgNode, err := FindByOrgPath(ctx, m.OrgPath)
	if err != nil {
		return fmt.Errorf("error finding org path: %w", err)
	}
	if orgNode == nil {
		return fmt.Errorf("org path not found: %s", m.OrgPath)
	}

	// 2. Check if Position exists
	position, err := FindPositionByKeyandPath(ctx, m.PositionKey, m.OrgPath)
	if err != nil {
		return fmt.Errorf("error finding position: %w", err)
	}
	if position == nil {
		return fmt.Errorf("position not found for key=%s and org_path=%s", m.PositionKey, m.OrgPath)
	}

	// 3. Create membership
	m.ID = bson.NewObjectID()
	m.CreatedAt = time.Now()
	m.UpdatedAt = time.Now()
	m.Active = true // optional: auto set active when created

	_, err = database.DB.Collection("memberships").InsertOne(ctx, m)
	if err != nil {
		return fmt.Errorf("error inserting membership: %w", err)
	}

	return nil
}

func GetUserMemberships(ctx context.Context, uid string) ([]models.Membership, error) {
	col := database.DB.Collection("memberships")

	filter := bson.M{
		"user_id" : uid,
		"active"  : true,
	}

	cursor, err := col.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var memberships []models.Membership
	if err := cursor.All(ctx, &memberships); err != nil {
		return nil, err
	}

	return memberships, nil
}

func (r *MembershipRepository) Delete(ctx context.Context, id any) error {
	_, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
