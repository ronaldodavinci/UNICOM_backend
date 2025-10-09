package repository

import (
	"context"
	"fmt"
	"main-webbase/internal/cursor"
	"main-webbase/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type CommentRepository struct {
	Client      *mongo.Client
	ColComments *mongo.Collection
	ColPosts    *mongo.Collection
}

// Create: เพิ่มคอมเมนต์ใหม่ + $inc comment_count (transaction)
func (r *CommentRepository) Create(ctx context.Context, postID, userID bson.ObjectID, text string) (*models.Comment, error) {
	now := time.Now().UTC()
	doc := &models.Comment{
		ID:        bson.NewObjectID(),
		PostID:    postID,
		UserID:    userID,
		Text:      text,
		CreatedAt: now,
		UpdatedAt: now,
		LikeCount: 0,
	}

	sess, err := r.Client.StartSession()
	if err != nil {
		return nil, err
	}
	defer sess.EndSession(ctx)

	_, err = sess.WithTransaction(ctx, func(sc context.Context) (interface{}, error) {
		if _, err := r.ColComments.InsertOne(sc, doc); err != nil {
			return nil, err
		}
		res, err := r.ColPosts.UpdateOne(
			sc,
			bson.M{"_id": postID},
			bson.M{"$inc": bson.M{"comment_count": 1}},
		)
		if err != nil {
			return nil, err
		}
		if res.MatchedCount == 0 {
			return nil, fmt.Errorf("post not found")
		}
		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return doc, nil
}

// ListByPostNewestFirst: รายการคอมเมนต์ของโพสต์ (ใหม่ก่อน) + cursor-based pagination
func (r *CommentRepository) ListByPostNewestFirst(
	ctx context.Context,
	postID bson.ObjectID,
	cursorStr string,
	limit int64,
) (items []models.Comment, next *string, err error) {

	// 1) base filter
	filter := bson.M{"post_id": postID}

	// 2) apply cursor ถ้ามี
	if cursorStr != "" {
		t, oid, derr := cursor.DecodeCommentCursor(cursorStr)
		if derr != nil {
			// ส่ง error ที่สื่อความหมายให้ handler map เป็น 400 ได้
			err = fmt.Errorf("invalid cursor: %w", derr)
			return
		}
		filter["$or"] = []bson.M{
			{"created_at": bson.M{"$lt": t}},
			{"created_at": t, "_id": bson.M{"$lt": oid}},
		}
	}

	// 3) sort + limit+1
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}, {Key: "_id", Value: -1}}).
		SetLimit(limit + 1)

	cur, err := r.ColComments.Find(ctx, filter, opts)
	if err != nil {
		return
	}
	defer cur.Close(ctx)

	var all []models.Comment
	if err = cur.All(ctx, &all); err != nil {
		return
	}

	// 4) ตัดหน้า และตั้ง next cursor
	if int64(len(all)) > limit {
		items = all[:limit]
		last := items[len(items)-1]
		s := cursor.EncodeCommentCursor(last.CreatedAt, last.ID)
		next = &s
	} else {
		items = all
		next = nil
	}
	return
}

// Update: แก้ไขข้อความคอมเมนต์ (เจ้าของหรือ root)
func (r *CommentRepository) Update(ctx context.Context, commentID, userID bson.ObjectID, newText string, isRoot bool) (*models.Comment, error) {
	var c models.Comment
	if err := r.ColComments.FindOne(ctx, bson.M{"_id": commentID}).Decode(&c); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	if c.UserID != userID && !isRoot {
		return nil, nil
	}

	now := time.Now().UTC()
	_, err := r.ColComments.UpdateOne(ctx,
		bson.M{"_id": commentID},
		bson.M{"$set": bson.M{"text": newText, "updated_at": now}},
	)
	if err != nil {
		return nil, err
	}

	c.Text = newText
	c.UpdatedAt = now
	return &c, nil
}

// Delete: ลบคอมเมนต์ (เจ้าของหรือ root) + $inc comment_count (transaction)
func (r *CommentRepository) Delete(ctx context.Context, commentID, userID bson.ObjectID, isRoot bool) (bool, error) {
	filter := bson.M{"_id": commentID}
	if !isRoot {
		filter["user_id"] = userID
	}

	sess, err := r.Client.StartSession()
	if err != nil {
		return false, err
	}
	defer sess.EndSession(ctx)

	deleted := false
	_, err = sess.WithTransaction(ctx, func(sc context.Context) (interface{}, error) {
		var c models.Comment
		if err := r.ColComments.FindOneAndDelete(sc, filter).Decode(&c); err != nil {
			if err == mongo.ErrNoDocuments {
				return nil, nil // ไม่เจอ/ไม่มีสิทธิ์ → ไม่ลบ
			}
			return nil, err
		}
		deleted = true

		// ✅ ใช้ pipeline update: comment_count = max(0, (comment_count || 0) - 1)
		update := mongo.Pipeline{
			bson.D{{Key: "$set", Value: bson.D{
				{Key: "comment_count", Value: bson.D{
					{Key: "$max", Value: bson.A{
						0,
						bson.D{{Key: "$subtract", Value: bson.A{
							bson.D{{Key: "$ifNull", Value: bson.A{"$comment_count", 0}}},
							1,
						}}},
					}},
				}},
			}}},
		}

		_, err := r.ColPosts.UpdateOne(
			sc,
			bson.M{"_id": c.PostID},
			update,
		)
		return nil, err
	})
	if err != nil {
		return false, err
	}

	return deleted, nil
}
