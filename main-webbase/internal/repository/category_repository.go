package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func ListAllCategories(ctx context.Context, client *mongo.Client) ([]bson.M, error) {
	coll := client.Database("unicom").Collection("categories")

	cur, err := coll.Find(
		ctx,
		bson.D{}, // no filter
		options.Find().
			SetProjection(bson.D{
				{Key: "_id", Value: 1},
				{Key: "category_name", Value: 1},
				{Key: "short_name", Value: 1},
			}).
			SetSort(bson.D{{Key: "category_name", Value: 1}}),
	)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var out []bson.M
	for cur.Next(ctx) {
		var m bson.M
		if err := cur.Decode(&m); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
