package repository

import (
	"context"
	"errors"

	"like_workspace/model"
	"like_workspace/internal/utils"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func InsertHashtags(db *mongo.Database, post model.Post, text string, ctx context.Context) error {
	hashtagsCol := db.Collection("hashtags")
	hashtags := utils.ExtractHashtags(text)
	if len(hashtags) == 0 {
		return nil
	}
	dateOnly := post.CreatedAt.Format("2006-01-02")
	docs := make([]interface{}, 0, len(hashtags))
	for _, tag := range hashtags {
		docs = append(docs, model.PostHashtag{
			PostID: post.ID,
			Tag:    tag,
			Date:   dateOnly,
		})
	}
	_, err := hashtagsCol.InsertMany(ctx, docs, options.InsertMany().SetOrdered(false))
	return err
}

func InsertCategories(db *mongo.Database, postID bson.ObjectID, categoryIDs []string, ctx context.Context) error {
	col := db.Collection("post_categories")
	docs := make([]interface{}, 0, len(categoryIDs))
	for i, cidStr := range categoryIDs {
		cid, err := bson.ObjectIDFromHex(cidStr)
		if err != nil {
			return errors.New("invalid categoryId: " + cidStr)
		}
		docs = append(docs, model.PostCategory{
			PostID:     postID,
			CategoryID: cid,
			OrderIndex: i + 1,
		})
	}
	_, err := col.InsertMany(ctx, docs, options.InsertMany().SetOrdered(false))
	return err
}

func InsertRoleVisibility(db *mongo.Database, postID bson.ObjectID, roleIDs []string, ctx context.Context) error {
	col := db.Collection("post_role_visibility")
	docs := make([]interface{}, 0, len(roleIDs))
	for _, ridStr := range roleIDs {
		rid, err := bson.ObjectIDFromHex(ridStr)
		if err != nil {
			return errors.New("invalid roleId: " + ridStr)
		}
		docs = append(docs, model.PostRoleVisibility{
			PostID: postID,
			RoleID: rid,
		})
	}
	_, err := col.InsertMany(ctx, docs, options.InsertMany().SetOrdered(false))
	return err
}
