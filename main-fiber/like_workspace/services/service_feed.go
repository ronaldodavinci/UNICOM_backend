package services

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"like_workspace/model"
	"like_workspace/internal/repository"
)

type Service struct {
	col *mongo.Collection
}

func NewService(col *mongo.Collection) *Service {
	return &Service{col: col}
}

func (s *Service) List(ctx context.Context, opt model.QueryOptions) ([]model.Post, *bson.ObjectID, error) {
	if opt.Limit <= 0 || opt.Limit > 100 {
		opt.Limit = 20
	}

	// ใช้ filter จาก feedquery (ใน repository)
	filter := repository.Build(repository.Options{
		Roles:      opt.Roles,
		Categories: opt.Categories,
		Tags:       opt.Tags,
		TextSearch: opt.TextSearch,
		SinceID:    opt.SinceID,
		UntilID:    opt.UntilID,
	})

	findOpt := options.Find().
		SetSort(bson.M{"_id": -1}).
		SetLimit(opt.Limit + 1)

	cur, err := s.col.Find(ctx, filter, findOpt)
	if err != nil {
		return nil, nil, err
	}
	defer cur.Close(ctx)

	var items []model.Post
	if err := cur.All(ctx, &items); err != nil {
		return nil, nil, err
	}

	var next *bson.ObjectID
	if int64(len(items)) == opt.Limit+1 {
		last := items[len(items)-1].ID
		items = items[:len(items)-1]
		next = &last
	}
	return items, next, nil
}
