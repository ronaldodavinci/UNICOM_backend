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
type PositionRepository struct {
    col *mongo.Collection
}

func NewPositionRepository() *PositionRepository {
    return &PositionRepository{
        col: config.DB.Collection("positions"),
    }
}

func (r *PositionRepository) Insert(ctx context.Context, p models.Position) error {
    p.ID = primitive.NewObjectID()
    p.CreatedAt = time.Now()
    p.UpdatedAt = time.Now()
    _, err := r.col.InsertOne(ctx, p)
    return err
}

func (r *PositionRepository) FindAll(ctx context.Context) ([]models.Position, error) {
    cur, err := r.col.Find(ctx, bson.M{})
    if err != nil {
        return nil, err
    }
    defer cur.Close(ctx)

    var positions []models.Position
    if err := cur.All(ctx, &positions); err != nil {
        return nil, err
    }
    return positions, nil
}