// internal/repository/posts_repository.go
package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Post = bson.M

// FetchPostsThatExistInPostCategories
// ดึงเฉพาะโพสต์ที่มีเอนทรีในคอลเลคชัน post_categories (pc ไม่ว่าง)
func FetchPostsThatExistInPostCategories(ctx context.Context, client *mongo.Client) ([]Post, error) {
	col := client.Database("lll_workspace").Collection("posts")

	// db.posts.aggregate([
	//   { $lookup: { from: "post_categories", localField: "_id", foreignField: "post_id", as: "pc" } },
	//   { $match:  { pc: { $ne: [] } } }
	// ])

	pipeline := mongo.Pipeline{
		{{
			Key: "$lookup",
			Value: bson.D{
				{Key: "from", Value: "post_categories"},
				{Key: "localField", Value: "_id"},
				{Key: "foreignField", Value: "post_id"},
				{Key: "as", Value: "pc"},
			},
		}},
		{{
			Key: "$match",
			// "pc": { $ne: [] }
			Value: bson.D{
				{Key: "pc", Value: bson.D{{Key: "$ne", Value: bson.A{}}}},
			},
		}},
		{{ Key: "$sort",  Value: bson.D{{Key: "created_at", Value: -1}, {Key: "_id", Value: -1}} }},
		{{ Key: "$limit", Value: 10 }},

		{{
		  Key: "$project",
		  Value: bson.D{
		    {Key: "pc", Value: 0},
		  },
		}},
		
	}

	// เผื่อ set timeout ให้ context (เลือกได้)
	cctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cur, err := col.Aggregate(cctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cur.Close(cctx)

	var results []Post
	if err := cur.All(cctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func FetchPostsWantedCategories(ctx context.Context, client *mongo.Client, catID primitive.ObjectID) ([]Post, error) {
	col := client.Database("lll_workspace").Collection("post_categories")

	// db.posts.aggregate([
	//   { $lookup: { from: "post_categories", localField: "_id", foreignField: "post_id", as: "pc" } },
	//   { $match:  { pc: { $ne: [] } } }
	// ])

	pipeline := mongo.Pipeline{
		{{
			Key: "$match",
			Value: bson.D{
				{Key: "category_id", Value: catID},
			},
		}},
		{{
			Key: "$lookup",
			Value: bson.D{
				{Key: "from", Value: "posts"},
				{Key: "localField", Value: "post_id"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "post"},
			},
		}},
		{{Key: "$unwind", Value: "$post"}}, // ได้เอกสารโพสต์ตรง ๆ
		{{
			Key: "$group", // กันโพสต์ซ้ำ
			Value: bson.D{
				{Key: "_id", Value: "$post._id"},
				{Key: "post", Value: bson.D{{Key: "$first", Value: "$post"}}},
				{Key: "matched_category_ids", Value: bson.D{{Key: "$addToSet", Value: "$category_id"}}},
			},
		}},
		{{
			Key: "$replaceRoot",
			Value: bson.D{
				{Key: "newRoot", Value: bson.D{
					{Key: "$mergeObjects", Value: bson.A{
						"$post",
						bson.D{{Key: "matched_category_ids", Value: "$matched_category_ids"}},
					}},
				}},
			},
		}},
		// เติมได้ภายหลัง:
		{{ Key: "$sort",  Value: bson.D{{Key: "created_at", Value: -1}, {Key: "_id", Value: -1}} }},
		{{ Key: "$limit", Value: 10 }},
		{{
		  Key: "$project",
		  Value: bson.D{
		    {Key: "pc", Value: 0},
		  },
		}},
	}

	// เผื่อ set timeout ให้ context (เลือกได้)
	cctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cur, err := col.Aggregate(cctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cur.Close(cctx)

	var results []Post
	if err := cur.All(cctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}