// internal/repository/hashtag_trending_repository.go
package repository

import (
	"context"
	"fmt"
	"strings"
	"time"
	"main-webbase/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// ================================
// Interface
// ================================
type HashtagTrendingRepository interface {
	// แบบที่ 1
	TopPublicHashtagsToday(ctx context.Context, k int) ([]bson.M, error)
	// แบบที่ 2
	TopPublicHashtagsAllTime(ctx context.Context, k int) ([]bson.M, error)
	// แบบที่ 2.1
	CountPublicPostsByHashtag(ctx context.Context, rawTag, day string) (*SingleTagCount, error)
	// แบบที่ 3
	ListPublicPostsByHashtag(ctx context.Context, rawTag, day string, limit int, cursorHex string) ([]models.Post, *bson.ObjectID, error)

	EnsureIndexes(ctx context.Context) error
}

// ================================
// Models
// ================================
type SingleTagCount struct {
	Tag       string `json:"tag" bson:"tag"`
	PostCount int    `json:"postCount" bson:"postCount"`
	Context   string `json:"context,omitempty" bson:"context,omitempty"`
}

// ================================
// Implementation (Mongo)
// ================================
type mongoHashtagTrendingRepo struct {
	db        *mongo.Database
	hashtags  *mongo.Collection
	posts     *mongo.Collection
	tzAsiaBkk *time.Location
}

// NOTE: ปรับให้ตรงกับที่คุณใช้อยู่ (รับ client และผูก DB ชื่อ unicom)
func NewMongoHashtagTrendingRepoWithDBName(client *mongo.Client) HashtagTrendingRepository {
	loc, _ := time.LoadLocation("Asia/Bangkok")
	db := client.Database("unicom")
	return &mongoHashtagTrendingRepo{
		db:        db,
		hashtags:  db.Collection("hashtags"),
		posts:     db.Collection("posts"),
		tzAsiaBkk: loc,
	}
}

// ---- helpers ----
func (r *mongoHashtagTrendingRepo) todayStr() string {
	now := time.Now().In(r.tzAsiaBkk)
	y, m, d := now.Date()
	return fmt.Sprintf("%04d-%02d-%02d", y, int(m), d) // "YYYY-MM-DD"
}

func (r *mongoHashtagTrendingRepo) normalizeInputTag(raw string) string {
	t := strings.ToLower(strings.TrimSpace(raw))
	if t == "" {
		return t
	}
	if !strings.HasPrefix(t, "#") {
		t = "#" + t
	}
	return t
}

// normalize tag (toLower + ensure '#') สำหรับ aggregation
func normTagExpr(input any) bson.M {
	return bson.M{
		"$cond": []any{
			bson.M{"$regexMatch": bson.M{"input": input, "regex": "^#"}},
			bson.M{"$toLower": input},
			bson.M{"$concat": []any{"#", bson.M{"$toLower": input}}},
		},
	}
}

// เงื่อนไขโพสต์สาธารณะ (รองรับ 2 สคีม่า) + optional active
// NOTE: ยอมรับกรณีไม่มีฟิลด์ visibility = public (default)
func buildPublicMatch(requireActive bool) bson.M {
	and := []bson.M{
		{
			"$or": []bson.M{
				{"p.visibility": "public"},                 // schema เก่า
				{"p.visibility.access": "public"},          // schema ใหม่
				{"p.visibility": bson.M{"$exists": false}}, // ไม่มีฟิลด์ = public
			},
		},
	}
	if requireActive {
		and = append(and, bson.M{"p.status": "active"})
	}
	return bson.M{"$and": and}
}

// ================================
// Queries
// ================================

// แบบที่ 1: Top today
func (r *mongoHashtagTrendingRepo) TopPublicHashtagsToday(ctx context.Context, k int) ([]bson.M, error) {
	if k <= 0 {
		k = 10
	}
	day := r.todayStr()

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"date": day}}},
		{{Key: "$lookup", Value: bson.M{
			"from": "posts",
			"let":  bson.M{"pid": "$postId"},
			"pipeline": mongo.Pipeline{
				{{Key: "$match", Value: bson.M{"$expr": bson.M{"$eq": bson.A{"$_id", "$$pid"}}}}},
				{{Key: "$match", Value: bson.M{"$and": []bson.M{
					{"$or": []bson.M{
						{"visibility": "public"},
						{"visibility.access": "public"},
						{"visibility": bson.M{"$exists": false}},
					}},
					{"status": "active"},
				}}}},
				{{Key: "$project", Value: bson.M{"_id": 1}}},
			},
			"as": "p",
		}}},
		{{Key: "$match", Value: bson.M{"$expr": bson.M{"$gt": bson.A{bson.M{"$size": "$p"}, 0}}}}},
		{{Key: "$addFields", Value: bson.M{"tagNorm": normTagExpr("$tag")}}},
		{{Key: "$group", Value: bson.M{"_id": bson.M{"post": "$postId", "tag": "$tagNorm"}}}},
		{{Key: "$group", Value: bson.M{"_id": "$_id.tag", "posts": bson.M{"$sum": 1}}}},
		{{Key: "$sort", Value: bson.M{"posts": -1, "_id": 1}}},
		{{Key: "$limit", Value: k}},
		{{Key: "$project", Value: bson.M{"_id": 0, "tag": "$_id", "postCount": "$posts", "context": "Trending"}}},
	}

	cur, err := r.hashtags.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var out []bson.M
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// แบบที่ 2: Top all-time
func (r *mongoHashtagTrendingRepo) TopPublicHashtagsAllTime(ctx context.Context, k int) ([]bson.M, error) {
	if k <= 0 {
		k = 10
	}

	pipeline := mongo.Pipeline{
		{{Key: "$lookup", Value: bson.M{
			"from": "posts",
			"let":  bson.M{"pid": "$postId"},
			"pipeline": mongo.Pipeline{
				{{Key: "$match", Value: bson.M{"$expr": bson.M{"$eq": bson.A{"$_id", "$$pid"}}}}},
				{{Key: "$match", Value: bson.M{"$and": []bson.M{
					{"$or": []bson.M{
						{"visibility": "public"},
						{"visibility.access": "public"},
						{"visibility": bson.M{"$exists": false}},
					}},
					{"status": "active"},
				}}}},
				{{Key: "$project", Value: bson.M{"_id": 1}}},
			},
			"as": "p",
		}}},
		{{Key: "$match", Value: bson.M{"$expr": bson.M{"$gt": bson.A{bson.M{"$size": "$p"}, 0}}}}},
		{{Key: "$addFields", Value: bson.M{"tagNorm": normTagExpr("$tag")}}},
		{{Key: "$group", Value: bson.M{"_id": bson.M{"post": "$postId", "tag": "$tagNorm"}}}},
		{{Key: "$group", Value: bson.M{"_id": "$_id.tag", "posts": bson.M{"$sum": 1}}}},
		{{Key: "$sort", Value: bson.M{"posts": -1, "_id": 1}}},
		{{Key: "$limit", Value: k}},
		{{Key: "$project", Value: bson.M{"_id": 0, "tag": "$_id", "postCount": "$posts", "context": "Trending"}}},
	}

	cur, err := r.hashtags.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var out []bson.M
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// แบบที่ 2.1: Count single tag
func (r *mongoHashtagTrendingRepo) CountPublicPostsByHashtag(ctx context.Context, rawTag, day string) (*SingleTagCount, error) {
	norm := r.normalizeInputTag(rawTag)
	if norm == "" {
		return nil, fmt.Errorf("tag is required")
	}

	match := bson.M{}
	if day != "" {
		match["date"] = day
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: match}},
		{{Key: "$addFields", Value: bson.M{"tagNorm": normTagExpr("$tag")}}},
		{{Key: "$match", Value: bson.M{"tagNorm": norm}}},
		{{Key: "$lookup", Value: bson.M{
			"from":         "posts",
			"localField":   "postId",
			"foreignField": "_id",
			"as":           "p",
		}}},
		{{Key: "$unwind", Value: "$p"}},
		{{Key: "$match", Value: buildPublicMatch(true)}},
		{{Key: "$group", Value: bson.M{"_id": "$postId"}}},
		{{Key: "$count", Value: "postCount"}},
	}

	cur, err := r.hashtags.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	count := 0
	if cur.Next(ctx) {
		var doc struct {
			PostCount int `bson:"postCount"`
		}
		if err := cur.Decode(&doc); err == nil {
			count = doc.PostCount
		}
	}
	return &SingleTagCount{Tag: norm, PostCount: count}, nil
}

// ================================
// Indexes
// ================================
func (r *mongoHashtagTrendingRepo) EnsureIndexes(ctx context.Context) error {
	_, err := r.hashtags.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "date", Value: 1}, {Key: "tag", Value: 1}}, Options: options.Index().SetName("date_tag")},
		{Keys: bson.D{{Key: "postId", Value: 1}}, Options: options.Index().SetName("postId")},
		{Keys: bson.D{{Key: "postId", Value: 1}, {Key: "tag", Value: 1}, {Key: "date", Value: 1}}, Options: options.Index().SetUnique(true).SetName("uniq_post_tag_day")},
	})
	if err != nil {
		return err
	}

	_, err = r.posts.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "visibility.access", Value: 1}, {Key: "created_at", Value: 1}}, Options: options.Index().SetName("visibilityAccess_createdAt")},
	})
	return err
}

// ================================
// แบบที่ 3: รายละเอียดโพสต์ (เหมือนฟีด)
// ================================
func (r *mongoHashtagTrendingRepo) ListPublicPostsByHashtag(
	ctx context.Context,
	rawTag, day string,
	limit int,
	cursorHex string,
) ([]models.Post, *bson.ObjectID, error) {

	// --- validate & prepare ---
	norm := r.normalizeInputTag(rawTag)
	if norm == "" {
		return nil, nil, fmt.Errorf("tag is required")
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	var cursorID *bson.ObjectID
	if cursorHex != "" {
		if oid, err := bson.ObjectIDFromHex(cursorHex); err == nil {
			cursorID = &oid
		}
	}

	// --- filter hashtag rows (collection: hashtags) ---
	match := bson.M{}
	if day != "" {
		match["date"] = day
	}

	base := mongo.Pipeline{
		bson.D{{Key: "$match", Value: match}},
		bson.D{{Key: "$addFields", Value: bson.M{"tagNorm": normTagExpr("$tag")}}},
		bson.D{{Key: "$match", Value: bson.M{"tagNorm": norm}}},
		bson.D{{Key: "$group", Value: bson.M{"_id": "$postId"}}}, // unique post ids
	}

	// --- join to posts, page by cursor (_id) ---
	itemsPipe := mongo.Pipeline{
		bson.D{{Key: "$lookup", Value: bson.M{
			"from":         "posts",
			"localField":   "_id",
			"foreignField": "_id",
			"as":           "p",
		}}},
		bson.D{{Key: "$unwind", Value: "$p"}},
	}
	if cursorID != nil {
		itemsPipe = append(itemsPipe, bson.D{{Key: "$match", Value: bson.M{"p._id": bson.M{"$lt": *cursorID}}}})
	}

	// --- จากนี้ทำเหมือนฟีด: replaceRoot -> posts แล้วทำ lookups/projection เดิม ---
	itemsPipe = append(itemsPipe,
		// ใช้เอกสาร p (post) เป็นราก
		bson.D{{Key: "$replaceRoot", Value: bson.M{"newRoot": "$p"}}},
		// สถานะ
		bson.D{{Key: "$match", Value: bson.M{"status": "active"}}},
		// ===== Users =====
		bson.D{{Key: "$lookup", Value: bson.M{
			"from":         "users",
			"localField":   "user_id",
			"foreignField": "_id",
			"as":           "u",
		}}},
		bson.D{{Key: "$unwind", Value: bson.M{"path": "$u", "preserveNullAndEmptyArrays": true}}},
		// ===== Visibility records =====
		bson.D{{Key: "$lookup", Value: bson.M{
			"from":         "post_role_visibility",
			"localField":   "_id",
			"foreignField": "post_id",
			"as":           "visRoles",
		}}},
		// ===== posted_as lookups =====
		bson.D{{Key: "$lookup", Value: bson.M{
			"from":         "org_units",
			"localField":   "node_id",
			"foreignField": "_id",
			"as":           "orgNode",
		}}},
		bson.D{{Key: "$unwind", Value: bson.M{"path": "$orgNode", "preserveNullAndEmptyArrays": true}}},
		bson.D{{Key: "$lookup", Value: bson.M{
			"from":         "positions",
			"localField":   "position_id",
			"foreignField": "_id",
			"as":           "pos",
		}}},
		bson.D{{Key: "$unwind", Value: bson.M{"path": "$pos", "preserveNullAndEmptyArrays": true}}},

		// ===== Categories =====
		bson.D{{Key: "$lookup", Value: bson.M{
			"from": "post_categories",
			"let":  bson.M{"pid": "$_id"},
			"pipeline": mongo.Pipeline{
				bson.D{{Key: "$match", Value: bson.M{"$expr": bson.M{"$eq": bson.A{"$post_id", "$$pid"}}}}},
				bson.D{{Key: "$sort", Value: bson.M{"order_index": 1}}},
			},
			"as": "pcAll",
		}}},
		bson.D{{Key: "$lookup", Value: bson.M{
			"from":         "categories",
			"localField":   "pcAll.category_id",
			"foreignField": "_id",
			"as":           "catAll",
		}}},
		bson.D{{Key: "$lookup", Value: bson.M{
			"from":         "comments",
			"localField":   "_id",
			"foreignField": "post_id",
			"as":           "comments",
		}}},

		// ===== Visibility helpers (เหมือนฟีด) =====
		bson.D{{Key: "$addFields", Value: bson.M{
			"hasVisibility": bson.M{"$gt": bson.A{
				bson.M{"$size": bson.M{"$ifNull": bson.A{"$visRoles", bson.A{}}}},
				0,
			}},
		}}},
		bson.D{{Key: "$addFields", Value: bson.M{
			"visibilityAccess": bson.M{"$cond": bson.A{"$hasVisibility", "private", "public"}},
		}}},

		// --- Projection ให้ตรง models.Post ---
		bson.D{{Key: "$project", Value: bson.M{
			"_id":        1,
			"user_id":    1,
			"name": bson.M{"$concat": bson.A{
				bson.M{"$ifNull": bson.A{"$u.firstname", ""}},
				" ",
				bson.M{"$ifNull": bson.A{"$u.lastname", ""}},
			}},
			"node_id":      1,
			"position_id":  1,
			"hashtag":      1,
			"tag":          1,
			"category": bson.M{
				"$map": bson.M{
					"input": bson.M{"$ifNull": bson.A{"$catAll", bson.A{}}}, "as": "c",
					"in":   bson.M{"$ifNull": bson.A{"$$c.category_name", ""}},
				},
			},
			"post_text": bson.M{"$ifNull": bson.A{"$censored_text", "$post_text"}},
			"media":         1,
			"like_count":    1,
			"comment_count": 1,
			"created_at":    1,
			"updated_at":    1,
			"status":        bson.M{"$ifNull": bson.A{"$status", "active"}},
			"visibility": bson.M{
				"$cond": bson.A{"$hasVisibility", "private", "public"},
			},
		}}},

		// --- sort & limit แบบฟีด ---
		bson.D{{Key: "$sort", Value: bson.M{"created_at": -1, "_id": -1}}},
		bson.D{{Key: "$limit", Value: limit + 1}},
	)

	// รวม pipeline (hashtags → posts like feed)
	pipeline := append(base,
		bson.D{{Key: "$facet", Value: bson.M{
			"items": itemsPipe,
		}}},
	)

	cur, err := r.hashtags.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, nil, err
	}
	defer cur.Close(ctx)

	var faceted []struct {
		Items []models.Post `bson:"items"`
	}
	if err := cur.All(ctx, &faceted); err != nil {
		return nil, nil, err
	}
	if len(faceted) == 0 {
		return []models.Post{}, nil, nil
	}

	items := faceted[0].Items

	var next *bson.ObjectID
	if len(items) == limit+1 {
		last := items[len(items)-1].ID
		items = items[:len(items)-1]
		next = &last
	}

	return items, next, nil
}
