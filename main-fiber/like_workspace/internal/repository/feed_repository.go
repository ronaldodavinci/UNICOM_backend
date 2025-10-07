// internal/repository/feed_repository.go
package repository

import (
	"context"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/Software-eng-01204341/Backend/model"
)

// ================================
// Interface
// ================================
type FeedRepository interface {
	List(ctx context.Context, opts model.QueryOptions) ([]model.FrontPost, *bson.ObjectID, error)
	// ใหม่: เรียงตามความนิยม (จำนวนไลค์มาก → น้อย)
	ListPopular(ctx context.Context, opts model.QueryOptions) ([]model.FrontPost, *bson.ObjectID, error)
}

// ================================
// Struct
// ================================
type mongoFeedRepo struct {
	col *mongo.Collection
}

func NewMongoFeedRepo(client *mongo.Client) FeedRepository {
	return &mongoFeedRepo{
		col: client.Database("lll_workspace").Collection("posts"),
	}
}

// ================================
// ฟังก์ชันสร้าง pipeline ส่วนกลาง
// ================================
func buildCommonPipeline(baseMatch bson.D, lim int64, opts model.QueryOptions, popularityMode bool) mongo.Pipeline {
	pipe := mongo.Pipeline{
		{{Key: "$match", Value: baseMatch}},
		// ===== Users =====
		{{Key: "$lookup", Value: bson.M{
			"from":         "users",
			"localField":   "user_id",
			"foreignField": "_id",
			"as":           "u",
		}}},
		{{Key: "$unwind", Value: bson.M{"path": "$u", "preserveNullAndEmptyArrays": true}}},
		// ===== Visibility records =====
		{{Key: "$lookup", Value: bson.M{
			"from":         "post_role_visibility",
			"localField":   "_id",
			"foreignField": "post_id",
			"as":           "visRoles",
		}}},
		// ===== posted_as lookups =====
		{{Key: "$lookup", Value: bson.M{
			"from":         "org_unit_node",
			"localField":   "role_path",
			"foreignField": "_id",
			"as":           "orgNode",
		}}},
		{{Key: "$unwind", Value: bson.M{"path": "$orgNode", "preserveNullAndEmptyArrays": true}}},
		{{Key: "$lookup", Value: bson.M{
			"from":         "positions",
			"localField":   "position_key",
			"foreignField": "_id",
			"as":           "pos",
		}}},
		{{Key: "$unwind", Value: bson.M{"path": "$pos", "preserveNullAndEmptyArrays": true}}},
		// ===== Categories =====
		{{Key: "$lookup", Value: bson.M{
			"from": "post_categories",
			"let":  bson.M{"pid": "$_id"},
			"pipeline": mongo.Pipeline{
				{{Key: "$match", Value: bson.M{"$expr": bson.M{"$eq": bson.A{"$post_id", "$$pid"}}}}},
				{{Key: "$sort", Value: bson.M{"order_index": 1}}},
			},
			"as": "pcAll",
		}}},
		{{Key: "$lookup", Value: bson.M{
			"from":         "categories",
			"localField":   "pcAll.category_id",
			"foreignField": "_id",
			"as":           "catAll",
		}}},
	}

	// ===== Text search =====
	if opts.TextSearch != "" {
		pipe = append(pipe, bson.D{{Key: "$match", Value: bson.M{"$or": []bson.M{
			{"post_text": bson.M{"$regex": opts.TextSearch, "$options": "i"}},
			{"u.username": bson.M{"$regex": opts.TextSearch, "$options": "i"}},
			{"u.user_firstname": bson.M{"$regex": opts.TextSearch, "$options": "i"}},
			{"u.user_lastname": bson.M{"$regex": opts.TextSearch, "$options": "i"}},
		}}}})
	}

	// ===== Role filters =====
	if len(opts.Roles) > 0 {
		orRole := make([]bson.M, 0, len(opts.Roles)*2)
		for _, r := range opts.Roles {
			r = strings.TrimSpace(r)
			if r == "" {
				continue
			}
			if strings.HasPrefix(r, "/") {
				// path or prefix
				if strings.HasSuffix(r, "/*") {
					prefix := strings.TrimSuffix(r, "/*")
					re := "^" + regexp.QuoteMeta(prefix)
					orRole = append(orRole, bson.M{"orgNode.path": bson.M{"$regex": re}})
				} else {
					orRole = append(orRole, bson.M{"orgNode.path": r})
				}
			} else {
				re := "^" + regexp.QuoteMeta(r) + "$"
				orRole = append(orRole, bson.M{"pos.key": bson.M{"$regex": re, "$options": "i"}})
				orRole = append(orRole, bson.M{"pos.name": bson.M{"$regex": re, "$options": "i"}})
			}
		}
		if len(orRole) > 0 {
			pipe = append(pipe, bson.D{{Key: "$match", Value: bson.M{"$or": orRole}}})
		}
	}

	// ===== Category filters =====
	if len(opts.Categories) > 0 {
		orCats := make([]bson.M, 0, len(opts.Categories)*2)
		for _, c := range opts.Categories {
			c = strings.TrimSpace(c)
			if c == "" {
				continue
			}
			re := "^" + regexp.QuoteMeta(c) + "$"
			orCats = append(orCats, bson.M{"catAll.category_name": bson.M{"$regex": re, "$options": "i"}})
			orCats = append(orCats, bson.M{"catAll.short_name": bson.M{"$regex": re, "$options": "i"}})
		}
		if len(orCats) > 0 {
			pipe = append(pipe, bson.D{{Key: "$match", Value: bson.M{"$or": orCats}}})
		}
	}

	// ===== Visibility helpers =====
	pipe = append(pipe,
		bson.D{{Key: "$addFields", Value: bson.M{
			"hasVisibility": bson.M{"$gt": bson.A{
				bson.M{"$size": bson.M{"$ifNull": bson.A{"$visRoles", bson.A{}}}},
				0,
			}},
		}}},
		bson.D{{Key: "$addFields", Value: bson.M{
			"visibilityAccess": bson.M{
				"$cond": bson.A{"$hasVisibility", "private", "public"},
			},
		}}},
		bson.D{{Key: "$addFields", Value: bson.M{
			"isOwner": bson.M{"$cond": bson.A{
				bson.M{"$ne": bson.A{opts.ViewerID, bson.ObjectID{}}},
				bson.M{"$eq": bson.A{"$user_id", opts.ViewerID}},
				false,
			}},
		}}},
	)

	if len(opts.AllowedNodeIDs) > 0 {
		pipe = append(pipe,
			bson.D{{Key: "$addFields", Value: bson.M{
				"allowedByNode": bson.M{
					"$gt": bson.A{
						bson.M{"$size": bson.M{"$filter": bson.M{
							"input": bson.M{"$ifNull": bson.A{"$visRoles", bson.A{}}}, "as": "vr",
							"cond": bson.M{"$in": bson.A{"$$vr.node_id", opts.AllowedNodeIDs}},
						}}},
						0,
					},
				},
			}}})
	} else {
		pipe = append(pipe, bson.D{{Key: "$addFields", Value: bson.M{"allowedByNode": false}}})
	}

	pipe = append(pipe,
		bson.D{{Key: "$match", Value: bson.M{"$or": []bson.M{
			{"hasVisibility": false},
			{"allowedByNode": true},
			{"isOwner": true},
		}}}},
	)

	// ===== Projection =====
	pipe = append(pipe,
		bson.D{{Key: "$project", Value: bson.M{
			"_id":     1,
			"user_id": 1,
			"uid":      bson.M{"$toString": "$u._id"},
			"username": bson.M{"$ifNull": bson.A{"$u.username", ""}},
			"name": bson.M{"$concat": bson.A{
				bson.M{"$ifNull": bson.A{"$u.user_firstname", ""}},
				" ",
				bson.M{"$ifNull": bson.A{"$u.user_lastname", ""}},
			}},
			"message":   "$post_text",
			"timestamp": "$created_at",
			"likes":     bson.M{"$ifNull": bson.A{"$like_count", 0}},
			"likedBy":   bson.A{},
			"posted_as": bson.M{
				"org_path":     bson.M{"$ifNull": bson.A{"$orgNode.path", ""}},
				"position_key": bson.M{"$ifNull": bson.A{"$pos.key", bson.M{"$ifNull": bson.A{"$pos.name", ""}}}},
			},
			"audience": bson.M{
				"$map": bson.M{
					"input": bson.M{"$ifNull": bson.A{"$catAll", bson.A{}}},
					"as":    "c",
					"in":    bson.M{"$ifNull": bson.A{"$$c.category_name", ""}},
				},
			},
			"visibility": bson.M{"access": "$visibilityAccess"},
			"org_of_content": bson.M{"$ifNull": bson.A{"$orgNode.path", ""}},
			"status":         bson.M{"$ifNull": bson.A{"$status", "active"}},
			"created_at":     "$created_at",
			"updated_at":     "$updated_at",
			"like_count":     bson.M{"$ifNull": bson.A{"$like_count", 0}},
		}}},
	)

	if popularityMode {
		pipe = append(pipe, bson.D{{Key: "$sort", Value: bson.M{"like_count": -1, "created_at": -1, "_id": -1}}})
	} else {
		pipe = append(pipe, bson.D{{Key: "$sort", Value: bson.M{"created_at": -1, "_id": -1}}})
	}
	pipe = append(pipe, bson.D{{Key: "$limit", Value: lim + 1}})
	return pipe
}

// ================================
// ListPopular (เรียงตามยอดไลค์มากสุด → น้อย)
// ================================
func (r *mongoFeedRepo) ListPopular(ctx context.Context, opts model.QueryOptions) ([]model.FrontPost, *bson.ObjectID, error) {
	lim := opts.Limit
	if lim <= 0 { lim = 20 }
	if lim > 20 { lim = 20 }

	var untilLike *int
	var untilTime *time.Time
	if !opts.UntilID.IsZero() {
		var tmp struct {
			LikeCount int       `bson:"like_count"`
			CreatedAt time.Time `bson:"created_at"`
		}
		_ = r.col.FindOne(
			ctx,
			bson.M{"_id": opts.UntilID},
			options.FindOne().SetProjection(bson.M{"like_count": 1, "created_at": 1, "_id": 0}),
		).Decode(&tmp)
		l := tmp.LikeCount
		untilLike = &l
		if !tmp.CreatedAt.IsZero() {
			t := tmp.CreatedAt.UTC()
			untilTime = &t
		}
	}

	baseMatch := bson.D{}
	if !opts.UntilID.IsZero() && untilLike != nil && untilTime != nil {
		baseMatch = append(baseMatch, bson.E{
			Key: "$or",
			Value: []bson.M{
				{"like_count": bson.M{"$lt": *untilLike}},
				{"like_count": *untilLike, "created_at": bson.M{"$lt": *untilTime}},
				{"like_count": *untilLike, "created_at": *untilTime, "_id": bson.M{"$lt": opts.UntilID}},
			},
		})
	} else if !opts.UntilID.IsZero() {
		baseMatch = append(baseMatch, bson.E{Key: "_id", Value: bson.M{"$lt": opts.UntilID}})
	}

	if len(opts.AuthorIDs) > 0 {
		baseMatch = append(baseMatch, bson.E{Key: "user_id", Value: bson.M{"$in": opts.AuthorIDs}})
	}

	pipe := buildCommonPipeline(baseMatch, lim, opts, true)

	cur, err := r.col.Aggregate(ctx, pipe, options.Aggregate())
	if err != nil { return nil, nil, err }
	defer cur.Close(ctx)

	var items []model.FrontPost
	if err := cur.All(ctx, &items); err != nil {
		return nil, nil, err
	}

	var next *bson.ObjectID
	if int64(len(items)) == lim+1 {
		last := items[len(items)-1].ID
		items = items[:len(items)-1]
		next = &last
	}
	return items, next, nil
}

// ================================
// List (เรียงตามเวลาล่าสุด → เก่าสุด)
// ================================
func (r *mongoFeedRepo) List(ctx context.Context, opts model.QueryOptions) ([]model.FrontPost, *bson.ObjectID, error) {
	var untilTime *time.Time
	var sinceTime *time.Time

	if !opts.UntilID.IsZero() {
		var tmp struct{ CreatedAt time.Time `bson:"created_at"` }
		_ = r.col.FindOne(ctx, bson.M{"_id": opts.UntilID}, options.FindOne().SetProjection(bson.M{"created_at": 1, "_id": 0})).Decode(&tmp)
		if !tmp.CreatedAt.IsZero() {
			t := tmp.CreatedAt.UTC()
			untilTime = &t
		}
	}
	if !opts.SinceID.IsZero() {
		var tmp struct{ CreatedAt time.Time `bson:"created_at"` }
		_ = r.col.FindOne(ctx, bson.M{"_id": opts.SinceID}, options.FindOne().SetProjection(bson.M{"created_at": 1, "_id": 0})).Decode(&tmp)
		if !tmp.CreatedAt.IsZero() {
			t := tmp.CreatedAt.UTC()
			sinceTime = &t
		}
	}

	baseMatch := bson.D{}
	if !opts.UntilID.IsZero() && untilTime != nil {
		baseMatch = append(baseMatch, bson.E{
			Key: "$or",
			Value: []bson.M{
				{"created_at": bson.M{"$lt": *untilTime}},
				{"created_at": *untilTime, "_id": bson.M{"$lt": opts.UntilID}},
			},
		})
	} else if !opts.SinceID.IsZero() && sinceTime != nil {
		baseMatch = append(baseMatch, bson.E{
			Key: "$or",
			Value: []bson.M{
				{"created_at": bson.M{"$gt": *sinceTime}},
				{"created_at": *sinceTime, "_id": bson.M{"$gt": opts.SinceID}},
			},
		})
	} else if !opts.UntilID.IsZero() {
		baseMatch = append(baseMatch, bson.E{Key: "_id", Value: bson.M{"$lt": opts.UntilID}})
	} else if !opts.SinceID.IsZero() {
		baseMatch = append(baseMatch, bson.E{Key: "_id", Value: bson.M{"$gt": opts.SinceID}})
	}

	if len(opts.AuthorIDs) > 0 {
		baseMatch = append(baseMatch, bson.E{Key: "user_id", Value: bson.M{"$in": opts.AuthorIDs}})
	}

	lim := opts.Limit
	if lim <= 0 { lim = 20 }
	if lim > 20 { lim = 20 }

	pipe := buildCommonPipeline(baseMatch, lim, opts, false)

	cur, err := r.col.Aggregate(ctx, pipe, options.Aggregate())
	if err != nil { return nil, nil, err }
	defer cur.Close(ctx)

	var items []model.FrontPost
	if err := cur.All(ctx, &items); err != nil {
		return nil, nil, err
	}

	var next *bson.ObjectID
	if int64(len(items)) == lim+1 {
		last := items[len(items)-1].ID
		items = items[:len(items)-1]
		next = &last
	}

	return items, next, nil
}
