package repositories

import (
	"context"

	"github.com/pllus/main-fiber/tamarind/config"
	"github.com/pllus/main-fiber/tamarind/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository struct {
	col *mongo.Collection
}

func NewUserRepository() *UserRepository {
	return &UserRepository{col: config.DB.Collection("users")}
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.col.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) UpdateRolesSummary(ctx context.Context, userID any, update bson.M) error {
	_, err := r.col.UpdateByID(ctx, userID, bson.M{"$set": update})
	return err
}
