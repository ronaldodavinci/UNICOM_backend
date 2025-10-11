package bootstrap

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func EnsureLikeIndexes(db *mongo.Database) error {
	_, err := db.Collection("like").Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "post_id", Value: 1},
				{Key: "comment_id", Value: 1},
			},
			Options: options.Index().SetUnique(true).SetName("uniq_user_post_comment"),
		},
	)
	return err
}
