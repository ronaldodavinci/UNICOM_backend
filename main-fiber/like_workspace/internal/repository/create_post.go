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
		err := nodeCol.FindOne(ctx, bson.M{"path": aud, "status": "active"}).Decode(&nodeDoc)
		if err != nil {
			return errors.New("org_path not found: " + aud)
		}

		doc := model.PostRoleVisibility{
			PostID: postID,
			NodeID: &nodeDoc.ID, // pointer เพื่อเก็บ ObjectID
		}
		docs = append(docs, doc)
	}

	_, err := col.InsertMany(ctx, docs, options.InsertMany().SetOrdered(false))
	return err
}

func FindUserInfo(col *mongo.Collection, userID bson.ObjectID, ctx context.Context) (user dto.UserInfoResponse , err error) {
	err = col.FindOne(ctx, bson.M{"_id": userID, "status": "active"}).Decode(&user)
	return user, err
}

// GetIndividualPostDetail
func FindPostByID(col *mongo.Collection, id bson.ObjectID, ctx context.Context) (model.Post, error) {
	var p model.Post
	err := col.FindOne(ctx, bson.M{"_id": id}).Decode(&p)
	return p, err
}
func FindPositionName(col *mongo.Collection, id bson.ObjectID, ctx context.Context) (string, error) {
	var doc struct {
		Name string `bson:"name"`
	}
	err := col.FindOne(ctx, bson.M{"_id": id, "status": "active"}).Decode(&doc)
	if err != nil {
		return "", err
	}
	return doc.Name, nil
}
func FindOrgNode(col *mongo.Collection, id bson.ObjectID, ctx context.Context) (string, error) {
	var doc struct {
		OrgPath string `bson:"path"`
	}
	err := col.FindOne(ctx, bson.M{"_id": id, "status": "active"}).Decode(&doc)
	return doc.OrgPath, err
}
// ดึง visibility ของโพสต์จาก post_role_visibility -> แปลง role_id เป็น org_unit_node.path
func FindVisibilityPaths(
    colPRV *mongo.Collection,       // post_role_visibility
    colOrg *mongo.Collection,       // org_unit_node
    postID bson.ObjectID,
    ctx context.Context,
) (dto.Visibility, error) {

    // 1) หา role_id ทั้งหมดที่ผูกกับ post_id
    cur, err := colPRV.Find(ctx,
        bson.M{"post_id": postID},
        options.Find().SetProjection(bson.M{"role_id": 1}))
    if err != nil {
        return dto.Visibility{}, err
    }
    defer cur.Close(ctx)

    roleIDs := make([]bson.ObjectID, 0, 8)
    for cur.Next(ctx) {
        var row struct {
            RoleID bson.ObjectID `bson:"role_id"`
        }
        if err := cur.Decode(&row); err != nil {
            return dto.Visibility{}, err
        }
        roleIDs = append(roleIDs, row.RoleID)
    }
    if err := cur.Err(); err != nil {
        return dto.Visibility{}, err
    }

    // ถ้าไม่มี record → access=public, audience=[]
    if len(roleIDs) == 0 {
        return dto.Visibility{Access: "public", Audience: []string{}}, nil
    }

    // 2) มี record → access=private แล้ว resolve path จาก org_unit_node
    //    (สมมติ role_id == _id ของ org_unit_node; ถ้า schema จริงต่างออกไป ให้ปรับ join ตรงนี้)
    cur2, err := colOrg.Find(ctx,
        bson.M{"_id": bson.M{"$in": roleIDs}, "status": "active"},
        options.Find().SetProjection(bson.M{"path": 1}))
    if err != nil {
        return dto.Visibility{}, err
    }
    defer cur2.Close(ctx)

    // เก็บ path แบบ unique
    pathSet := make(map[string]struct{}, len(roleIDs))
    for cur2.Next(ctx) {
        var node struct {
            Path string `bson:"path"`
        }
        if err := cur2.Decode(&node); err != nil {
            return dto.Visibility{}, err
        }
        if node.Path != "" {
            pathSet[node.Path] = struct{}{}
        }
    }
    if err := cur2.Err(); err != nil {
        return dto.Visibility{}, err
    }

    audience := make([]string, 0, len(pathSet))
    for p := range pathSet {
        audience = append(audience, p)
    }

    return dto.Visibility{Access: "private", Audience: audience}, nil
}
// ดึง category_ids ของโพสต์จาก post_categories
func FindCategoryIDs(col *mongo.Collection, postID bson.ObjectID, ctx context.Context) ([]string, error) {
	cur, err := col.Find(ctx, bson.M{"post_id": postID}, options.Find().SetProjection(bson.M{"category_id": 1}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var ids []string
	for cur.Next(ctx) {
		var row struct {
			CategoryID bson.ObjectID `bson:"category_id"`
		}
		if err := cur.Decode(&row); err != nil {
			return nil, err
		}
		ids = append(ids, row.CategoryID.Hex())
	}
	return ids, cur.Err()
}


// DeletePostHandler
func DeletePost(db *mongo.Database, postID bson.ObjectID, ctx context.Context) error {
	col := db.Collection("posts")

	// filter หาจาก _id
	filter := bson.M{"_id": postID, "status": "active"}
	update := bson.M{
		"$set": bson.M{
			"status": "inactive",
		},
	}

	result, err := col.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("post not found or already inactive")
	}

	return nil
}
