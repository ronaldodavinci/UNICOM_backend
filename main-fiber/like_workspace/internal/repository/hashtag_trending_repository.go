// internal/repository/hashtag_trending_repository.go
package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

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
	ListPublicPostsByHashtag(ctx context.Context, rawTag, day string, limit int, cursorHex string) ([]bson.M, *bson.ObjectID, error)

	EnsureIndexes(ctx context.Context) error
}

// ================================
// Models
// ================================
type SingleTagCount struct {
	Tag       string `json:"tag" bson:"tag"`
	PostCount int    `json:"postCount" bson:"postCount"`
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

// NOTE: ปรับให้ตรงกับที่คุณใช้อยู่ (รับ client และผูก DB ชื่อ lll_workspace)
func NewMongoHashtagTrendingRepoWithDBName(client *mongo.Client) HashtagTrendingRepository {
	loc, _ := time.LoadLocation("Asia/Bangkok")
	db := client.Database("lll_workspace")
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
				{"p.visibility": "public"},              // schema เก่า
				{"p.visibility.access": "public"},       // schema ใหม่
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
		{{Key: "$project", Value: bson.M{"_id": 0, "tag": "$_id", "postCount": "$posts"}}},
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
		{{Key: "$project", Value: bson.M{"_id": 0, "tag": "$_id", "postCount": "$posts"}}},
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
		var doc struct{ PostCount int `bson:"postCount"` }
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
// แบบที่ 3: รายละเอียดโพสต์ (with cursor + facet total)
// ================================
func (r *mongoHashtagTrendingRepo) ListPublicPostsByHashtag(
	ctx context.Context,
	rawTag, day string,
	limit int,
	cursorHex string,
) ([]bson.M, *bson.ObjectID, error) {

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

	match := bson.M{}
	if day != "" {
		match["date"] = day
	}

	base := mongo.Pipeline{
		bson.D{{Key: "$match", Value: match}},
		bson.D{{Key: "$addFields", Value: bson.M{"tagNorm": normTagExpr("$tag")}}},
		bson.D{{Key: "$match", Value: bson.M{"tagNorm": norm}}},
		bson.D{{Key: "$group", Value: bson.M{"_id": "$postId"}}},
	}

	itemsPipe := mongo.Pipeline{
		bson.D{{Key: "$lookup", Value: bson.M{
			"from":         "posts",
			"localField":   "_id",
			"foreignField": "_id",
			"as":           "p",
		}}},
		bson.D{{Key: "$unwind", Value: "$p"}},
		bson.D{{Key: "$match", Value: buildPublicMatch(true)}},
	}
	if cursorID != nil {
		itemsPipe = append(itemsPipe, bson.D{{Key: "$match", Value: bson.M{"p._id": bson.M{"$lt": *cursorID}}}})
	}
	itemsPipe = append(itemsPipe,
		bson.D{{Key: "$sort", Value: bson.M{"p.created_at": -1, "p._id": -1}}},
		bson.D{{Key: "$limit", Value: limit + 1}},
		bson.D{{Key: "$lookup", Value: bson.M{
			"from": "users",
			"let":  bson.M{"uidStr": "$p.uid", "userId": "$p.user_id"},
			"pipeline": mongo.Pipeline{
				bson.D{{Key: "$match", Value: bson.M{
					"$expr": bson.M{"$or": bson.A{
						bson.M{"$eq": bson.A{"$_id", "$$userId"}},
						bson.M{"$eq": bson.A{"$_id", bson.M{"$toObjectId": "$$uidStr"}}},
					}},
				}}},
				bson.D{{Key: "$project", Value: bson.M{"_id": 1, "name": 1, "username": 1}}},
			},
			"as": "u",
		}}},
		bson.D{{Key: "$unwind", Value: bson.M{"path": "$u", "preserveNullAndEmptyArrays": true}}},
		bson.D{{Key: "$addFields", Value: bson.M{"tag": norm}}},
		bson.D{{Key: "$project", Value: bson.M{
			"_id":       "$p._id",
			"uid":       bson.M{"$ifNull": bson.A{"$p.uid", bson.M{"$toString": "$p.user_id"}}},
			"username":  bson.M{"$ifNull": bson.A{"$p.username", bson.M{"$ifNull": bson.A{"$u.username", ""}}}},
			"name":      bson.M{"$ifNull": bson.A{"$p.name", bson.M{"$ifNull": bson.A{"$u.name", ""}}}},
			"message":   bson.M{"$ifNull": bson.A{"$p.message", bson.M{"$ifNull": bson.A{"$p.post_text", ""}}}},
			"timestamp": "$p.created_at",
			"likes":     bson.M{"$ifNull": bson.A{"$p.likes", 0}},
			"likedBy":   bson.M{"$ifNull": bson.A{"$p.likedBy", bson.A{}}},
			"posted_as": bson.M{"$ifNull": bson.A{"$p.posted_as", bson.M{}}},
			"tag":       1,
			"visibility": bson.M{"$ifNull": bson.A{
				"$p.visibility",
				bson.M{"access": "public", "org_of_content": ""},
			}},
		}}},
	)

	totalPipe := mongo.Pipeline{
		bson.D{{Key: "$lookup", Value: bson.M{
			"from":         "posts",
			"localField":   "_id",
			"foreignField": "_id",
			"as":           "p",
		}}},
		bson.D{{Key: "$unwind", Value: "$p"}},
		bson.D{{Key: "$match", Value: buildPublicMatch(true)}},
		bson.D{{Key: "$count", Value: "postCount"}},
	}

	pipeline := append(base,
		bson.D{{Key: "$facet", Value: bson.M{
			"items": itemsPipe,
			"total": totalPipe,
		}}},
	)

	cur, err := r.hashtags.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, nil, err
	}
	defer cur.Close(ctx)

	var faceted []struct {
		Items []bson.M `bson:"items"`
		Total []struct {
			PostCount int `bson:"postCount"`
		} `bson:"total"`
	}
	if err := cur.All(ctx, &faceted); err != nil {
		return nil, nil, err
	}
	if len(faceted) == 0 {
		return []bson.M{}, nil, nil
	}

	items := faceted[0].Items

	var next *bson.ObjectID
	if len(items) == limit+1 {
		last := items[len(items)-1]
		items = items[:len(items)-1]
		if oid, ok := last["_id"].(bson.ObjectID); ok {
			next = &oid
		}
	}

	return items, next, nil
}
