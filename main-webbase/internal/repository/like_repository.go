package repository

import (
	"context"
	"errors"
	"fmt"
	"main-webbase/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func InsertLike(ctx context.Context, likesCol *mongo.Collection, likeDoc models.Like) (dup bool, err error) {
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

func DeleteLike(ctx context.Context, likesCol *mongo.Collection, userID, targetID bson.ObjectID, targetType string) error {
	var filter bson.M
	if targetType == "post" {
		filter = bson.M{"user_id": userID, "post_id": targetID}
	} else {
		filter = bson.M{"user_id": userID, "comment_id": targetID}
	}
	_, err := likesCol.DeleteOne(ctx, filter)
	return err
}

func IncLikeCount(ctx context.Context, updateCol *mongo.Collection, targetFilter bson.M) (likeCount int64, err error) {
	_, err = updateCol.UpdateOne(ctx, targetFilter, bson.M{"$inc": bson.M{"like_count": 1}})

	return likeCount, err
}

func DecLikeCount(ctx context.Context, col *mongo.Collection, filter bson.M) (likeCount int64, err error) {
	_, err = col.UpdateOne(ctx, filter, bson.M{"$inc": bson.M{"like_count": -1}})

	return likeCount, err
}

func CountLikes(ctx context.Context, likesCol *mongo.Collection, targetID bson.ObjectID, targetType string) (int64, error) {
	var filter bson.M
	if targetType == "post" {
		filter = bson.M{"post_id": targetID}
	} else {
		filter = bson.M{"comment_id": targetID}
	}
	return likesCol.CountDocuments(ctx, filter)
}

func CheckIsLiked(
    ctx context.Context,
    col *mongo.Collection,
    userID bson.ObjectID,
    targetID bson.ObjectID,
    targetType string,
) (bool, error) {
    var filter bson.M

    switch targetType {
    case "post":
        filter = bson.M{
            "user_id": userID,
            "post_id": targetID,
        }
    case "comment":
        filter = bson.M{
            "user_id":    userID,
            "comment_id": targetID,
        }
    default:
        return false, fmt.Errorf("invalid targetType: must be 'post' or 'comment'")
    }

    count, err := col.CountDocuments(ctx, filter)
    if err != nil {
        return false, fmt.Errorf("failed to check like: %w", err)
    }

    return count > 0, nil
}