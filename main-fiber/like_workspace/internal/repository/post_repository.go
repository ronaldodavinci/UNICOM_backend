package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// FetchPostsVisible returns posts visible to a given user by role or ownership.
func FetchPostsVisible(
	ctx context.Context,
	client *mongo.Client,
	userID string,
	roles []string,
	limit int64,
) ([]map[string]interface{}, error) {

	coll := client.Database("lll_workspace").Collection("posts")

	pipeline := mongo.Pipeline{
		// Join role visibility
		{{Key: "$lookup", Value: bson.M{
			"from":         "post_rolevisibility",
			"localField":   "_id",
			"foreignField": "post_id",
			"as":           "rv",
		}}},
		// Extract rv_roles (array of roles) or fallback empty array
		{{Key: "$addFields", Value: bson.M{
			"rv_roles": bson.M{"$ifNull": bson.A{
				bson.M{"$let": bson.M{
					"vars": bson.M{"firstRv": bson.M{"$first": "$rv"}},
					"in":   bson.M{"$ifNull": bson.A{"$$firstRv.roles", bson.A{}}},
				}},
				bson.A{},
			}},
		}}},
		// Visibility rules:
		// - rv_roles empty (public)
		// - intersection with user roles > 0
		// - or author matches userID
		{{Key: "$match", Value: bson.M{
			"$expr": bson.M{
				"$or": bson.A{
					bson.M{"$eq": bson.A{bson.M{"$size": "$rv_roles"}, 0}},
					bson.M{"$gt": bson.A{
						bson.M{"$size": bson.M{"$setIntersection": bson.A{"$rv_roles", roles}}},
						0,
					}},
					bson.M{"$eq": bson.A{"$author_id", userID}},
				},
			},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "timestamp", Value: -1}}}},
		{{Key: "$limit", Value: limit}},
	}

	cursor, err := coll.Aggregate(ctx, pipeline, options.Aggregate())
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []map[string]interface{}
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}
