package repository

import (
	"context"
	"errors"

	"like_workspace/dto"
	"like_workspace/internal/utils"
	"like_workspace/model"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// แปลง org_path → node_id (ObjectID) จาก collection org_unit_node
func ResolveOrgNodeIDByPath(db *mongo.Database, orgPath string, ctx context.Context) (bson.ObjectID, error) {
	var doc struct {
		ID bson.ObjectID `bson:"_id"`
	}
	err := db.Collection("org_unit_node").FindOne(ctx, bson.M{"path": orgPath, "status": "active"}).Decode(&doc)
	return doc.ID, err
}

// แปลง position_key → pos_id (ObjectID) จาก collection positions
func ResolvePositionIDByKey(db *mongo.Database, positionKey string, ctx context.Context) (bson.ObjectID, error) {
	var doc struct {
		ID bson.ObjectID `bson:"_id"`
	}
	err := db.Collection("positions").FindOne(ctx, bson.M{"key": positionKey, "status": "active"}).Decode(&doc)
	return doc.ID, err
}

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

func InsertRoleVisibility(db *mongo.Database, postID bson.ObjectID, visibility dto.Visibility, ctx context.Context) error {
	col := db.Collection("post_role_visibility")

	audN := len(visibility.Audience)

	if audN == 0 {
		docs := model.PostRoleVisibility{
			PostID: postID,
			NodeID: nil,
			Scope:  "self",
		}
		_, err := col.InsertOne(ctx, docs)
		return err
	}

	// เตรียม map org_path → node_id จากตาราง org_unit_node
	nodeCol := db.Collection("org_unit_node")
	docs := make([]interface{}, 0, audN)

	for _, aud := range visibility.Audience {
		var nodeDoc struct {
			ID     bson.ObjectID `bson:"_id"`
			Path   string        `bson:"path"`
			Status string        `bson:"status"`
		}
		err := nodeCol.FindOne(ctx, bson.M{"path": aud.OrgPath, "status": "active"}).Decode(&nodeDoc)
		if err != nil {
			return errors.New("org_path not found: " + aud.OrgPath)
		}

		scope := aud.Scope

		doc := model.PostRoleVisibility{
			PostID: postID,
			NodeID: &nodeDoc.ID, // pointer เพื่อเก็บ ObjectID
			Scope:  scope,
		}
		docs = append(docs, doc)
	}

	_, err := col.InsertMany(ctx, docs, options.InsertMany().SetOrdered(false))
	return err
}

func FindUserInfo(db *mongo.Database, userID bson.ObjectID, ctx context.Context) (user dto.UserInfoResponse , err error) {
	col := db.Collection("users")
	err = col.FindOne(ctx, bson.M{"_id": userID, "status": "active"}).Decode(&user)
	return user, err
}
