package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"like_workspace/internal/utils"
)

// FetchPostsVisible returns posts visible to a given user by role or ownership.
// func FetchPostsVisible(
// 	ctx context.Context,
// 	client *mongo.Client,
// 	userID string,
// 	roles []string,
// 	limit int64,
// ) ([]map[string]interface{}, error) {

// 	coll := client.Database("lll_workspace").Collection("posts")

// 	pipeline := mongo.Pipeline{
// 	// 1) $lookup with pipeline (compare as strings to survive mixed types)
// 	{{Key: "$lookup", Value: bson.M{
// 		"from": "post_role_visibility", // <-- correct collection name
// 		"let":  bson.M{"pid": "$_id"},
// 		"pipeline": bson.A{
// 			bson.M{"$match": bson.M{
// 				"$expr": bson.M{
// 					"$eq": bson.A{
// 						bson.M{"$toString": "$post_id"},
// 						bson.M{"$toString": "$$pid"},
// 					},
// 				},
// 			}},
// 		},
// 		"as": "rv",
// 	}}},

// 	// 2) Extract all role IDs from joined rows -> rv_role_ids (unique array)
// 	{{Key: "$addFields", Value: bson.M{
// 		"rv_role_ids": bson.M{
// 			"$setUnion": bson.A{
// 				bson.M{"$ifNull": bson.A{
// 					bson.M{"$map": bson.M{
// 						"input": "$rv",
// 						"as":    "r",
// 						"in":    "$$r.role_id", // keep as-is; you can also normalize here if needed
// 					}},
// 					bson.A{},
// 				}},
// 				bson.A{},
// 			},
// 		},
// 	}}},

// 	// 3) Visibility: public (no rows) OR role-intersection OR owner
// 	{{Key: "$match", Value: bson.M{
// 		"$expr": bson.M{
// 			"$or": bson.A{
// 				bson.M{"$eq": bson.A{bson.M{"$size": "$rv_role_ids"}, 0}},
// 				// NOTE: 'roles' must be same type as rv_role_ids elements.
// 				bson.M{"$gt": bson.A{
// 					bson.M{"$size": bson.M{"$setIntersection": bson.A{"$rv_role_ids", roles}}},
// 					0,
// 				}},
// 				// Owner check (compare as strings to survive mixed types)
// 				bson.M{"$eq": bson.A{
// 					bson.M{"$toString": "$user_id"},
// 					bson.M{"$toString": userID},
// 				}},
// 			},
// 		},
// 	}}},

// 	{{Key: "$sort",  Value: bson.D{{Key: "created_at", Value: -1}}}},
// 	{{Key: "$limit", Value: limit}},
// }

// 	cursor, err := coll.Aggregate(ctx, pipeline, options.Aggregate())
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer cursor.Close(ctx)

// 	var results []map[string]interface{}
// 	if err := cursor.All(ctx, &results); err != nil {
// 		return nil, err
// 	}

// 	return results, nil
// }

// func FetchPostsAllPlainRaw(ctx context.Context, client *mongo.Client, limit int64) ([]bson.M, error) {
// 	col := client.Database("lll_workspace").Collection("posts")
// 	cur, err := col.Find(ctx, bson.M{}, options.Find().
// 		SetSort(bson.M{"created_at": -1}).
// 		SetLimit(limit))
// 	if err != nil { return nil, err }
// 	defer cur.Close(ctx)

// 	out := make([]bson.M, 0)
// 	if err := cur.All(ctx, &out); err != nil { return nil, err }
// 	return out, nil
// }

const createdAtField = "created_at"

func FetchPostsVisibleCursor(
    ctx context.Context,
    client *mongo.Client,
    userID string,
    roles []string,
    limit int64,
    after *time.Time,
) ([]bson.M, *time.Time, error) {

	coll := client.Database("lll_workspace").Collection("posts")

	pipeline := mongo.Pipeline{}

	// ถ้ามี cursor -> filter created_at < after (ทั้งหมดทำในโค้ด ไม่รับจากผู้ใช้)
	if after != nil {
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: bson.D{
			{Key: createdAtField, Value: bson.D{{Key: "$lt", Value: *after}}},
		}}})
	}

	// join สิทธิ์การมองเห็น
	pipeline = append(pipeline, bson.D{{Key: "$lookup", Value: bson.M{
		"from": "post_role_visibility",
		"let":  bson.M{"pid": "$_id"},
		"pipeline": bson.A{
			bson.M{"$match": bson.M{
				"$expr": bson.M{
					"$eq": bson.A{
						bson.M{"$toString": "$post_id"},
						bson.M{"$toString": "$$pid"},
					},
				},
			}},
		},
		"as": "rv",
	}}})

	// ดึง role_id เป็นอาเรย์
	pipeline = append(pipeline, bson.D{{Key: "$addFields", Value: bson.M{
		"rv_role_ids": bson.M{
			"$setUnion": bson.A{
				bson.M{"$ifNull": bson.A{
					bson.M{"$map": bson.M{
						"input": "$rv",
						"as":    "r",
						"in":    "$$r.role_id",
					}},
					bson.A{},
				}},
				bson.A{},
			},
		},
	}}})

	// เงื่อนไขการมองเห็น: public (ไม่มี row), intersect role, หรือเจ้าของ
	pipeline = append(pipeline, bson.D{{Key: "$match", Value: bson.M{
		"$expr": bson.M{
			"$or": bson.A{
				bson.M{"$eq": bson.A{bson.M{"$size": "$rv_role_ids"}, 0}},
				bson.M{"$gt": bson.A{
					bson.M{"$size": bson.M{"$setIntersection": bson.A{"$rv_role_ids", roles}}},
					0,
				}},
				bson.M{"$eq": bson.A{
					bson.M{"$toString": "$user_id"},
					bson.M{"$toString": userID},
				}},
			},
		},
	}}})

	// sort desc + ดึงเกินมา 1 เพื่อเช็คว่ามีหน้าถัดไปไหม
	pipeline = append(pipeline,
		bson.D{{Key: "$sort", Value: bson.D{{Key: createdAtField, Value: -1}}}},
		bson.D{{Key: "$limit", Value: limit + 1}},
	)

	cur, err := coll.Aggregate(ctx, pipeline, options.Aggregate())
	if err != nil {
		return nil, nil, err
	}
	defer cur.Close(ctx)

	var docs []bson.M
	if err := cur.All(ctx, &docs); err != nil {
		return nil, nil, err
	}

	var next *time.Time
	if int64(len(docs)) > limit {
		if t, ok := utils.ExtractTime(docs[limit], createdAtField); ok {
			next = &t
		}
		docs = docs[:limit]
	}
	return docs, next, nil
}
