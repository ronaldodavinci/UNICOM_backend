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

// EnsureEventParticipantIndexes enforces uniqueness of a participant per (event_id, user_id, role)
// to prevent duplicate rows when toggling form requirements or submitting forms multiple times.
func EnsureEventParticipantIndexes(db *mongo.Database) error {
    _, err := db.Collection("event_participant").Indexes().CreateOne(
        context.Background(),
        mongo.IndexModel{
            Keys: bson.D{
                {Key: "event_id", Value: 1},
                {Key: "user_id", Value: 1},
                {Key: "role", Value: 1},
            },
            Options: options.Index().SetUnique(true).SetName("uniq_event_user_role"),
        },
    )
    return err
}
