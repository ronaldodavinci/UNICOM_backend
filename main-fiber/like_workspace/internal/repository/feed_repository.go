package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"like_workspace/model"
)

// ===== MongoDB stage/keyword constants =====
const (
	StageMatch     = "$match"
	StageLookup    = "$lookup"
	StageUnwind    = "$unwind"
	StageAddFields = "$addFields"
	StageProject   = "$project"
	StageSort      = "$sort"
	StageLimit     = "$limit"

	KeyFrom         = "from"
	KeyLocalField   = "localField"
	KeyForeignField = "foreignField"
	KeyAs           = "as"
	KeyPipeline     = "pipeline"
	KeyLet          = "let"
)

// ===== Repository =====

type FeedRepository interface {
	List(ctx context.Context, opts model.QueryOptions) ([]model.FrontPost, *bson.ObjectID, error)
}

type mongoFeedRepo struct {
	col *mongo.Collection
}

func NewMongoFeedRepo(client *mongo.Client) FeedRepository {
	return &mongoFeedRepo{
		col: client.Database("lll_workspace").Collection("posts"),
	}
}


func BaseVisibilityFilter(roles []string, onlyPublic *bool, viewerID bson.ObjectID) bson.M {
	if onlyPublic != nil && *onlyPublic {
		if viewerID != (bson.ObjectID{}) {
			return bson.M{"$or": []bson.M{
				{"is_public": true},
				{"author_id": viewerID},
			}}
		}
		return bson.M{"is_public": true}
	}

	if len(roles) == 0 {
		if viewerID != (bson.ObjectID{}) {
			return bson.M{"$or": []bson.M{
				{"is_public": true},
				{"author_id": viewerID},
			}}
		}
		return bson.M{"is_public": true}
	}

	or := []bson.M{
		{"is_public": true},
		{"visibility.roles": bson.M{"$in": roles}},
	}
	if viewerID != (bson.ObjectID{}) {
		or = append(or, bson.M{"author_id": viewerID})
	}
	return bson.M{"$or": or}
}

func Build(opts model.QueryOptions) bson.M {
	and := []bson.M{BaseVisibilityFilter(opts.Roles, opts.OnlyPublic, opts.ViewerID)}

	if len(opts.Categories) > 0 {
		and = append(and, bson.M{"categories": bson.M{"$in": opts.Categories}})
	}
	if len(opts.Tags) > 0 {
		and = append(and, bson.M{"tags": bson.M{"$in": opts.Tags}})
	}
	if len(opts.AuthorIDs) > 0 {
		and = append(and, bson.M{"author_id": bson.M{"$in": opts.AuthorIDs}})
	}
	if opts.TextSearch != "" {
		and = append(and, bson.M{"$text": bson.M{"$search": opts.TextSearch}})
	}
	if opts.UntilID != (bson.ObjectID{}) {
		and = append(and, bson.M{"_id": bson.M{"$lt": opts.UntilID}})
	}
	if opts.SinceID != (bson.ObjectID{}) {
		and = append(and, bson.M{"_id": bson.M{"$gt": opts.SinceID}})
	}

	if len(and) == 1 {
		return and[0]
	}
	return bson.M{"$and": and}
}

// ใช้ opts โดยตรงให้แน่ใจว่า cursor ถูก apply แน่นอน
func adoptBaseMatchFromFilter(opts model.QueryOptions) bson.D {
	m := bson.D{}

	// ใส่เฉพาะหนึ่งอย่างตามทิศทางเพจ ถ้าใช้ until (เลื่อนลง) ก็ไม่ควรใส่ since พร้อมกัน
	if !opts.UntilID.IsZero() {
		m = append(m, bson.E{Key: "_id", Value: bson.M{"$lt": opts.UntilID}})
	} else if !opts.SinceID.IsZero() {
		m = append(m, bson.E{Key: "_id", Value: bson.M{"$gt": opts.SinceID}})
	}

	// filter ผู้เขียน (posts.user_id)
	if len(opts.AuthorIDs) > 0 {
		m = append(m, bson.E{Key: "user_id", Value: bson.M{"$in": opts.AuthorIDs}})
	}

	return m
}

func (r *mongoFeedRepo) List(ctx context.Context, opts model.QueryOptions) ([]model.FrontPost, *bson.ObjectID, error) {
	baseMatch := adoptBaseMatchFromFilter(opts)

	var viewerRoles []string
	if len(opts.Roles) > 0 {
		viewerRoles = opts.Roles
	}

	lim := opts.Limit
	if lim <= 0 {
		lim = 20
	}
	if lim > 100 {
		lim = 100
	}

	pipe := mongo.Pipeline{
		{{Key: StageMatch, Value: baseMatch}},

		// == Join ผู้ใช้ ==
		{{Key: StageLookup, Value: bson.M{
			KeyFrom:         "users",
			KeyLocalField:   "user_id",
			KeyForeignField: "_id",
			KeyAs:           "u",
		}}},
		{{Key: StageUnwind, Value: bson.M{"path": "$u", "preserveNullAndEmptyArrays": true}}},

		// == Join role ของผู้เขียน ==
		{{Key: StageLookup, Value: bson.M{
			KeyFrom:         "roles",
			KeyLocalField:   "role_id",
			KeyForeignField: "_id",
			KeyAs:           "authorRole",
		}}},
		{{Key: StageUnwind, Value: bson.M{"path": "$authorRole", "preserveNullAndEmptyArrays": true}}},

		// == Join visibility roles ของโพสต์ ==
		{{Key: StageLookup, Value: bson.M{
			KeyFrom:         "post_role_visibility",
			KeyLocalField:   "_id", // NOTE: โครงสร้างเดิมในโปรเจกต์ (ถ้า schema จริงคือ post_id ให้ปรับตรงนี้เป็น post_id)
			KeyForeignField: "post_id",
			KeyAs:           "visRoles",
		}}},
	}

	// == หมวดหมู่ (เลือก category แรก) ==
	pipe = append(pipe,
		bson.D{{Key: StageLookup, Value: bson.M{
			KeyFrom: "post_categories",
			KeyLet:  bson.M{"pid": "$_id"},
			KeyPipeline: mongo.Pipeline{
				{{Key: StageMatch, Value: bson.M{"$expr": bson.M{"$eq": bson.A{"$post_id", "$$pid"}}}}},
				{{Key: StageSort, Value: bson.M{"order_index": 1}}},
				{{Key: StageLimit, Value: 1}},
			},
			KeyAs: "pc",
		}}},
		bson.D{{Key: StageUnwind, Value: bson.M{"path": "$pc", "preserveNullAndEmptyArrays": true}}},
		bson.D{{Key: StageLookup, Value: bson.M{
			KeyFrom:         "categories",
			KeyLocalField:   "pc.category_id",
			KeyForeignField: "_id",
			KeyAs:           "cat",
		}}},
		bson.D{{Key: StageUnwind, Value: bson.M{"path": "$cat", "preserveNullAndEmptyArrays": true}}},
	)

	// == ค้นหาแบบ text/filters ==
	if opts.TextSearch != "" {
		pipe = append(pipe, bson.D{{Key: StageMatch, Value: bson.M{"$or": []bson.M{
			{"post_text": bson.M{"$regex": opts.TextSearch, "$options": "i"}},
			{"u.username": bson.M{"$regex": opts.TextSearch, "$options": "i"}},
			{"u.user_firstname": bson.M{"$regex": opts.TextSearch, "$options": "i"}},
			{"u.user_lastname": bson.M{"$regex": opts.TextSearch, "$options": "i"}},
		}}}})
	}
	if len(opts.Categories) > 0 {
		pipe = append(pipe, bson.D{{Key: StageMatch, Value: bson.M{"cat.category_id": bson.M{"$in": opts.Categories}}}})
	}
	if len(opts.Tags) > 0 {
		pipe = append(pipe, bson.D{{Key: StageMatch, Value: bson.M{"cat.category_name": bson.M{"$in": opts.Tags}}}})
	}

	// == คำนวณสถานะการมองเห็นพื้นฐาน ==
	pipe = append(pipe, bson.D{{Key: StageAddFields, Value: bson.M{
		"visibilityAccess": bson.M{
			"$cond": bson.A{
				bson.M{"$gt": bson.A{bson.M{"$size": "$visRoles"}, 0}},
				"private",          // มีรายการ role visibility → ไม่ใช่ public
				"public",    // ไม่มีรายการ role visibility → public
			},
		},
	}}})

	// == เพิ่ม isOwner เพื่อยกเว้นให้เจ้าของเห็นเสมอ ==
	pipe = append(pipe, bson.D{{Key: StageAddFields, Value: bson.M{
		"isOwner": bson.M{"$cond": bson.A{
			bson.M{"$ne": bson.A{opts.ViewerID, bson.ObjectID{}}}, // viewer มีตัวตนไหม
			bson.M{"$eq": bson.A{"$user_id", opts.ViewerID}},      // เป็นเจ้าของโพสต์ไหม
			false,
		}},
	}}})

	// == คำนวณ allowedByRole (ถ้าผู้ชมมี roles) แล้ว OR การมองเห็นกับ isOwner ==
	if len(viewerRoles) > 0 {
		pipe = append(pipe,
			bson.D{{Key: StageAddFields, Value: bson.M{
				"allowedByRole": bson.M{
					"$gt": bson.A{
						bson.M{"$size": bson.M{"$filter": bson.M{
							"input": "$visRoles",
							"as":    "vr",
							"cond":  bson.M{"$in": bson.A{"$$vr.role_id", viewerRoles}},
						}}},
						0,
					},
				},
			}}},
			bson.D{{Key: StageMatch, Value: bson.M{"$or": []bson.M{
				{"visibilityAccess": "public"},
				{"allowedByRole": true},
				{"isOwner": true}, // 👈 เจ้าของเห็นเสมอ
			}}}},
		)
	} else {
		// ไม่มี role → เห็นเฉพาะ public + เจ้าของโพสต์
		pipe = append(pipe, bson.D{{Key: StageMatch, Value: bson.M{"$or": []bson.M{
			{"visibilityAccess": "public"},
			{"isOwner": true}, // 👈 เจ้าของเห็นเสมอ
		}}}})
	}

	// == เลือกฟิลด์ส่งออก ==
	pipe = append(pipe,
		bson.D{{Key: StageProject, Value: bson.M{
			"_id":       1,
			"uid":       "$u.user_id",
			"name":      bson.M{"$concat": bson.A{"$u.user_firstname", " ", "$u.user_lastname"}},
			"username":  "$u.username",
			"message":   "$post_text",
			"timestamp": "$created_at",
			"likes":     bson.M{"$ifNull": bson.A{"$like_count", 0}},
			"likedBy":   bson.A{},
			"posted_as": bson.M{
				"org_path":     "$authorRole.role_path",
				"position_key": "$authorRole.role_id",
			},
			"tag": "$cat.category_name",
			"visibility": bson.M{
				"access":         "$visibilityAccess",
				"org_of_content": "$authorRole.role_path",
			},
		}}},

		bson.D{{Key: StageSort, Value: bson.M{"_id": -1}}},
		bson.D{{Key: StageLimit, Value: lim + 1}},
	)

	cur, err := r.col.Aggregate(ctx, pipe, options.Aggregate())
	if err != nil {
		return nil, nil, err
	}
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
