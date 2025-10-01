package repositories

import (
	"context"

	"github.com/pllus/main-fiber/tamarind/config/config"
	"github.com/pllus/main-fiber/tamarind/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type PostRepository struct {
	col *mongo.Collection
}

func NewPostRepository() *PostRepository {
	return &PostRepository{col: config.DB.Collection("posts")}
}

func (r *PostRepository) Insert(ctx context.Context, p models.Post) error {
	_, err := r.col.InsertOne(ctx, p)
	return err
}

func (r *PostRepository) FindAll(ctx context.Context) ([]models.Post, error) {
	cur, err := r.col.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	var posts []models.Post
	if err := cur.All(ctx, &posts); err != nil {
		return nil, err
	}
	return posts, nil
}
