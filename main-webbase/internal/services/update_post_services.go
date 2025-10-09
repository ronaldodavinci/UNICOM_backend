package services

import (
	"context"
	"fmt"
	"main-webbase/dto"

	"main-webbase/internal/models"
	repo "main-webbase/internal/repository"
	u "main-webbase/internal/utils"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func UpdatePostFull(
	client *mongo.Client,
	db *mongo.Database,
	postID, userID bson.ObjectID,
	isRoot bool,
	in dto.UpdatePostFullDTO,
	ctx context.Context,
) (*models.Post, error) {

	sess, err := client.StartSession()
	if err != nil {
		return nil, err
	}
	defer sess.EndSession(ctx)

	var updated *models.Post
	// fmt.Printf("[Svc] UpdatePostFull start post=%s isRoot=%v\n", postID.Hex(), isRoot)

	_, err = sess.WithTransaction(ctx, func(tx context.Context) (any, error) {
		// 1) resolve PostAs (เหมือน create)
		rolePathID, err := repo.ResolveOrgNodeIDByPath(db, in.PostAs.OrgPath, tx)
		if err != nil {
			return nil, fmt.Errorf("org_path not found4")
		}
		positionID, err := repo.ResolvePositionIDByKey(db, in.PostAs.PositionKey, tx)
		if err != nil {
			return nil, fmt.Errorf("position_key not found")
		}

		// fmt.Printf("[Svc2] UpdatePostFull start post=%s isRoot=%v\n", postID.Hex(), isRoot)
		// 2) update core post
		out, err := repo.UpdatePostCore(db, postID, userID, isRoot, in, rolePathID, positionID, tx)
		if err != nil {
			return nil, err
		}
		if out == nil {

			// fmt.Println("[FORBIDDEN-2] post not found or not owner")
			return nil, fmt.Errorf("forbidden or not found")
		}
		updated = out

		// 3) replace categories (ลบเก่า→ใส่ใหม่)
		if err := repo.ReplaceCategories(db, postID, in.CategoryIDs, tx); err != nil {
			return nil, err
		}

		// 4) replace visibility
		if in.Visibility.Access == "private" {
			if err := repo.ReplaceRoleVisibility(db, postID, in.Visibility, tx); err != nil {
				return nil, err
			}
		} else {
			// public → เคลียร์ทิ้ง
			if _, err := db.Collection("post_role_visibility").
				DeleteMany(tx, bson.M{"post_id": postID}); err != nil {
				return nil, err
			}
		}

		// 5) rebuild hashtags (เก็บทั้งใน posts และตาราง hashtags)
		updated.Hashtag = u.ExtractHashtags(in.PostText)
		if err := repo.RebuildHashtags(db, *updated, in.PostText, tx); err != nil {
			// จะ treat เป็น non-fatal ก็ได้ แต่ที่นี่ขอให้ fail ตรง ๆ เพื่อความสม่ำเสมอ
			return nil, err
		}

		return nil, nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}
