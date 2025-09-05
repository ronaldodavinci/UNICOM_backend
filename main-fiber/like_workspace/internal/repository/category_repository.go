// internal/repository/posts_repository.go
package repository

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Cursor payload ที่เราจะ encode เป็น base64
type pageCursor struct {
	T  time.Time          `json:"t"`  // created_at
	ID primitive.ObjectID `json:"id"` // _id
}

func encodeCursor(t time.Time, id primitive.ObjectID) (string, error) {
	b, err := json.Marshal(pageCursor{T: t, ID: id})
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func decodeCursor(s string) (pageCursor, error) {
	if s == "" {
		return pageCursor{}, errors.New("empty")
	}
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return pageCursor{}, err
	}
	var c pageCursor
	if err := json.Unmarshal(raw, &c); err != nil {
		return pageCursor{}, err
	}
	return c, nil
}

// helper: ดึงเวลาจาก bson.M ให้ได้ time.Time (รองรับทั้ง time.Time และ primitive.DateTime)
func timeFromAny(v any) (time.Time, bool) {
	switch x := v.(type) {
	case time.Time:
		return x, true
	case primitive.DateTime:
		return x.Time(), true
	}
	return time.Time{}, false
}

// FetchPostsWithCategoriesPage ดึงหน้าโพสต์ + categories แบบ cursor pagination (ไม่ใช้ struct)
// - sort: created_at desc, _id desc
// - cursor: base64(JSON{t,id}) ของ item สุดท้ายในหน้า
// รับพารามิเตอร์ฟิลเตอร์ด้วย
func FetchPostsWithCategoriesCursor(
	ctx context.Context,
	client *mongo.Client,
	limit int64,
	cursor string,
	catIDs []primitive.ObjectID, // ฟิลเตอร์ด้วย _id ของหมวด
	catNames []string,           // หรือฟิลเตอร์ด้วยชื่อหมวด
) (items []bson.M, nextCursor string, err error) {

	if limit <= 0 {
		limit = 10
	}
	limitPlusOne := limit + 1

	// ===== cursor pre-match =====
	match := bson.D{}
	if c, derr := decodeCursor(cursor); derr == nil {
		match = bson.D{{Key: "$or", Value: bson.A{
			bson.D{{Key: "created_at", Value: bson.D{{Key: "$lt", Value: c.T}}}},
			bson.D{
				{Key: "created_at", Value: c.T},
				{Key: "_id", Value: bson.D{{Key: "$lt", Value: c.ID}}},
			},
		}}}
	}

	pipeline := mongo.Pipeline{}
	if len(match) > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: match}})
	}

	// ===== ฟิลเตอร์ด้วย catIDs ก่อน sort/limit (ประหยัดและถูกต้องกับ cursor) =====
	if len(catIDs) > 0 {
		pipeline = append(pipeline,
			bson.D{
				{Key: "$lookup", Value: bson.D{
					{Key: "from", Value: "post_categories"},
					{Key: "let", Value: bson.D{{Key: "pid", Value: "$_id"}}},
					{Key: "pipeline", Value: bson.A{
						bson.D{
							{Key: "$match", Value: bson.D{
								{Key: "$expr", Value: bson.D{
									{Key: "$and", Value: bson.A{
										bson.D{{Key: "$eq", Value: bson.A{"$post_id", "$$pid"}}},
										bson.D{{Key: "$in", Value: bson.A{"$category_id", catIDs}}},
									}},
								}},
							}},
						},
						bson.D{
							{Key: "$limit", Value: 1},
						},
					}},
					{Key: "as", Value: "pc_hit"},
				}},
			},
			bson.D{
				{Key: "$match", Value: bson.D{
					{Key: "$expr", Value: bson.D{
						{Key: "$gt", Value: bson.A{
							bson.D{{Key: "$size", Value: "$pc_hit"}},
							0,
						}},
					}},
				}},
			},
			bson.D{
				{Key: "$unset", Value: "pc_hit"},
			},
		)
	}

	// ===== sort ตาม created_at/_id เพื่อรองรับ cursor =====
	pipeline = append(pipeline,
		bson.D{{Key: "$sort", Value: bson.D{
			{Key: "created_at", Value: -1},
			{Key: "_id", Value: -1},
		}}},
	)

	// ===== join เพื่อเตรียม categories (ต้องมาก่อน match ด้วยชื่อ) =====
	pipeline = append(pipeline,
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "post_categories"},
			{Key: "localField", Value: "_id"},
			{Key: "foreignField", Value: "post_id"},
			{Key: "as", Value: "pc"},
		}}},
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "cat_ids", Value: bson.D{{Key: "$ifNull", Value: bson.A{
				bson.D{{Key: "$map", Value: bson.D{
					{Key: "input", Value: "$pc"},
					{Key: "as", Value: "m"},
					{Key: "in", Value: "$$m.category_id"},
				}}},
				bson.A{},
			}}}},
		}}},
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "categories"},
			{Key: "let", Value: bson.D{{Key: "ids", Value: "$cat_ids"}}},
			{Key: "pipeline", Value: bson.A{
				bson.D{{Key: "$match", Value: bson.D{
					{Key: "$expr", Value: bson.D{{Key: "$in", Value: bson.A{"$_id", "$$ids"}}}},
				}}},
				bson.D{{Key: "$project", Value: bson.D{
					{Key: "_id", Value: 1},
					{Key: "name", Value: 1},
				}}},
				bson.D{{Key: "$sort", Value: bson.D{{Key: "name", Value: 1}}}},
			}},
			{Key: "as", Value: "categories"},
		}}},
	)

	// ===== ฟิลเตอร์ด้วยชื่อหมวด ต้องทำ "หลัง" มีฟิลด์ categories แล้ว และยังคงทำก่อน limit =====
	if len(catNames) > 0 {
		pipeline = append(pipeline,
			bson.D{{Key: "$match", Value: bson.D{
				{Key: "categories", Value: bson.D{{Key: "$in", Value: catNames}}},
			}}},
		)
	}

	// ===== สุดท้ายค่อย limit+1 เพื่อเช็กหน้าถัดไป =====
	pipeline = append(pipeline,
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 1},
			{Key: "post_text", Value: 1},
			{Key: "user_id", Value: 1},
			{Key: "role_id", Value: 1},
			{Key: "created_at", Value: 1},
			{Key: "categories", Value: 1},
		}}},
		bson.D{{Key: "$limit", Value: limitPlusOne}},
	)

	col := client.Database("lll_workspace").Collection("posts")
	cur, err := col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, "", err
	}
	defer cur.Close(ctx)

	var rows []bson.M
	if err := cur.All(ctx, &rows); err != nil {
		return nil, "", err
	}

	// มีหน้าถัดไปไหม
	if int64(len(rows)) > limit {
		last := rows[limit-1]
		oid, _ := last["_id"].(primitive.ObjectID)

		var ct time.Time
		if t, ok := timeFromAny(last["created_at"]); ok {
			ct = t
		}
		if nc, encErr := encodeCursor(ct, oid); encErr == nil {
			nextCursor = nc
		}
		rows = rows[:limit]
	}
	return rows, nextCursor, nil
}


// internal/repository/posts_by_category.go
type Post = bson.M

func FetchPostsByCategory(
	ctx context.Context,
	client *mongo.Client,
	categoryID string,    // รับเป็น string แล้วแปลงเป็น ObjectID
	limit int64,          // เผื่อจำกัดจำนวน
) ([]Post, error) {

	cid, err := primitive.ObjectIDFromHex(categoryID)
	if err != nil {
		return nil, err
	}

	pipeline := mongo.Pipeline{
		// (ถ้าข้อมูลใหญ่ แนะนำใช้เวอร์ชัน $lookup + pipeline แบบกรอง cid ตั้งแต่ใน lookup ดูหัวข้อ 3)
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "post_categories"},
			{Key: "localField", Value: "_id"},
			{Key: "foreignField", Value: "post_id"},
			{Key: "as", Value: "pc"},
		}}},
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "pc.category_id", Value: cid},
		}}},
		// sort ใหม่สุดก่อน เผื่อทำฟีด
		bson.D{{Key: "$sort", Value: bson.D{{Key: "created_at", Value: -1}, {Key: "_id", Value: -1}}}},
		// จำกัดฟิลด์ที่จะส่งออก (ปรับตามต้องการ)
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "pc", Value: 0}, // ไม่ส่ง pc ออกไปถ้าไม่จำเป็น
		}}},
	}
	if limit > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$limit", Value: limit}})
	}

	col := client.Database("lll_workspace").Collection("posts")
	cur, err := col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var out []Post
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}
