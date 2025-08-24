package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func FetchUsers(client *mongo.Client) ([]bson.M, error) {
	db := client.Database("User_1")
	collection := db.Collection("User")

	cursor, err := collection.Find(context.TODO(), bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var users []bson.M
	if err = cursor.All(context.TODO(), &users); err != nil {
		return nil, err
	}

	return users, nil
}
