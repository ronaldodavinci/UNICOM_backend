package repository

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"like_workspace/model"
)

// ย้ายบล็อก InsertOne เดิมมาไว้ที่นี่ และคงตรรกะ duplicate (11000)
func InsertLike(ctx context.Context, likesCol *mongo.Collection, likeDoc model.Like) (dup bool, err error) {
	_, err = likesCol.InsertOne(ctx, likeDoc)
	if err == nil {
		return false, nil
	}
	var we mongo.WriteException
	if errors.As(err, &we) && len(we.WriteErrors) > 0 && we.WriteErrors[0].Code == 11000 {
		return true, nil
	}
	return false, err
}

func IncLikeCount(ctx context.Context, updateCol *mongo.Collection, targetFilter bson.M) error {
	_, _ = updateCol.UpdateOne(ctx, targetFilter, bson.M{"$inc": bson.M{"like_count": 1}})
	return nil
}