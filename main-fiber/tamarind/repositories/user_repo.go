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
	return &UserRepository{
		col: config.DB.Collection("users"),
	}
}

func (r *UserRepository) Insert(ctx context.Context, u models.User) (*mongo.InsertOneResult, error) {
	return r.col.InsertOne(ctx, u)
}

func (r *UserRepository) FindAll(ctx context.Context) ([]models.User, error) {
	cur, err := r.col.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	var users []models.User
	if err := cur.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id interface{}) (models.User, error) {
	var u models.User
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&u)
	return u, err
}

func (r *UserRepository) Update(ctx context.Context, id interface{}, update interface{}) error {
	_, err := r.col.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": update})
	return err
}

func (r *UserRepository) Delete(ctx context.Context, id interface{}) error {
	_, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (models.User, error) {
	var u models.User
	err := r.col.FindOne(ctx, bson.M{"email": email}).Decode(&u)
	return u, err
}

func (r *UserRepository) UpdateRolesSummary(ctx context.Context, userID interface{}, update interface{}) error {
	_, err := r.col.UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{"$set": update},
	)
	return err
}