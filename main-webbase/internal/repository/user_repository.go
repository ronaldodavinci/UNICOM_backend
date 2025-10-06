package repository

import (
	"context"
	"time"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"main-webbase/database"
	"main-webbase/internal/models"
)

func FindUserBy(ctx context.Context, field, value string) ([]models.User, error) {
	collection := database.DB.Collection("users")

	var filter bson.M
	if field == "_id" {
		objID, err := bson.ObjectIDFromHex(value)
		if err != nil {
			return nil, err
		}
		filter = bson.M{"_id": objID}
	} else {
		filter = bson.M{field: value}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var users []models.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	if len(users) == 0 {
			return nil, fmt.Errorf("no users found")
	}

	return users, err
}