// internal/repository/posts_repository.go
package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"like_workspace/internal/cursor"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongo "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// ดึงโพสต์ทั้งหมดจาก posts พร้อม field visibility ("public"/"private")
// และรองรับกรองตาม visibility:
//   visibility == ""         -> ไม่กรอง (ทั้งหมด)
//   visibility == "public"   -> เอาเฉพาะสาธารณะ
//   visibility == "private"  -> เอาเฉพาะที่มี mapping ใน post_role_visibility
//
// หมายเหตุ:
// - ถ้า visibility ถูกกำหนด จะคำนวณ total เป็นจำนวนที่ตรงตามเงื่อนไขกรอง
// - ถ้า visibility ว่าง จะนับ total เป็นจำนวนโพสต์ทั้งหมดใน posts (เหมือนเดิม)
func ListAllPostsWithVisibilityNewestFirst(
	ctx context.Context,
	client *mongo.Client,
	cursorStr string,
	visibility string, // "", "public", "private"
	limit int64,
) (items []bson.M, next *string, total int64, err error) {

	db := client.Database("lll_workspace")
	postsColl := db.Collection("posts")

	vis := strings.ToLower(strings.TrimSpace(visibility))
	if vis != "" && vis != "public" && vis != "private" {
		return nil, nil, 0, fmt.Errorf("invalid visibility: %q (use \"public\", \"private\", or empty)", visibility)
	}

	// 1) total
	{
		cctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		if vis == "" {
			// นับทุกโพสต์ใน posts (พฤติกรรมเดิม)
			n, e := postsColl.CountDocuments(cctx, bson.D{})
			if e != nil {
				return nil, nil, 0, e
			}
			total = n
		} else {
			// นับตามเงื่อนไข visibility ด้วย aggregate
			countPipe := mongo.Pipeline{
				// join เพื่อตรวจว่าโพสต์มี prv หรือไม่
				bson.D{{Key: "$lookup", Value: bson.D{
					{Key: "from", Value: "post_role_visibility"},
					{Key: "localField", Value: "_id"},
					{Key: "foreignField", Value: "post_id"},
					{Key: "as", Value: "prv"},
				}}},
			}

			switch vis {
			case "private":
				// มี prv อย่างน้อย 1
				countPipe = append(countPipe, bson.D{{Key: "$match", Value: bson.D{
					{Key: "$expr", Value: bson.D{{Key: "$gt", Value: bson.A{bson.D{{Key: "$size", Value: "$prv"}}, 0}}}},
				}}})
			case "public":
				// ไม่มี prv เลย
				countPipe = append(countPipe, bson.D{{Key: "$match", Value: bson.D{
					{Key: "$expr", Value: bson.D{{Key: "$eq", Value: bson.A{bson.D{{Key: "$size", Value: "$prv"}}, 0}}}},
				}}})
			}

			countPipe = append(countPipe, bson.D{{Key: "$count", Value: "n"}})

			cur, e := postsColl.Aggregate(cctx, countPipe)
			if e != nil {
				return nil, nil, 0, e
			}
			var tmp []bson.M
			if e := cur.All(cctx, &tmp); e != nil {
				return nil, nil, 0, e
			}
			if len(tmp) > 0 {
				if v, ok := tmp[0]["n"].(int32); ok {
					total = int64(v)
				} else if v64, ok := tmp[0]["n"].(int64); ok {
					total = v64
				} else {
					total = 0
				}
			} else {
				total = 0
			}
		}
	}

	// 2) cursor match (created_at, _id) ใช้กับ posts
	var cursorMatch bson.D
	if cursorStr != "" {
		t, oid, e := cursor.DecodePostCursor(cursorStr)
		if e != nil {
			return nil, nil, 0, e
		}
		cursorMatch = bson.D{{
			Key: "$or",
			Value: bson.A{
				bson.D{{Key: "created_at", Value: bson.D{{Key: "$lt", Value: t}}}},
				bson.D{{Key: "created_at", Value: t}, {Key: "_id", Value: bson.D{{Key: "$lt", Value: oid}}}},
			},
		}}
	}

	// 3) pipeline หลัก
	pipe := mongo.Pipeline{}

	if len(cursorMatch) > 0 {
		pipe = append(pipe, bson.D{{Key: "$match", Value: cursorMatch}})
	}

	pipe = append(pipe,
		// เผื่อ created_at เป็น string -> แปลงเป็น Date ให้ sort/paginate ถูกต้อง
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "created_at", Value: bson.D{{Key: "$toDate", Value: "$created_at"}}},
		}}},
		// join หา prv
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "post_role_visibility"},
			{Key: "localField", Value: "_id"},
			{Key: "foreignField", Value: "post_id"},
			{Key: "as", Value: "prv"},
		}}},
	)

	// กรองตาม visibility ถ้าระบุ
	if vis == "private" {
		pipe = append(pipe, bson.D{{Key: "$match", Value: bson.D{
			{Key: "$expr", Value: bson.D{{Key: "$gt", Value: bson.A{bson.D{{Key: "$size", Value: "$prv"}}, 0}}}},
		}}})
	} else if vis == "public" {
		pipe = append(pipe, bson.D{{Key: "$match", Value: bson.D{
			{Key: "$expr", Value: bson.D{{Key: "$eq", Value: bson.A{bson.D{{Key: "$size", Value: "$prv"}}, 0}}}},
		}}})
	}

	// คำนวณ field ที่ต้องการ
	pipe = append(pipe,
		// สร้าง visibility + matched_role_ids
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "visibility", Value: bson.D{
				{Key: "$cond", Value: bson.A{
					bson.D{{Key: "$gt", Value: bson.A{bson.D{{Key: "$size", Value: "$prv"}}, 0}}},
					"private",
					"public",
				}},
			}},
			{Key: "matched_role_ids", Value: bson.D{
				{Key: "$cond", Value: bson.A{
					bson.D{{Key: "$gt", Value: bson.A{bson.D{{Key: "$size", Value: "$prv"}}, 0}}},
					bson.D{{Key: "$setUnion", Value: bson.A{"$prv.role_id", bson.A{}}}},
					bson.A{},
				}},
			}},
		}}},

		// กัน null
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "like_count", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$like_count", 0}}}},
			{Key: "comment_count", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$comment_count", 0}}}},
		}}},

		// ไม่ส่ง array prv ออกไป
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "prv", Value: 0},
		}}},

		// เรียงใหม่ -> เก่า
		bson.D{{Key: "$sort", Value: bson.D{{Key: "created_at", Value: -1}, {Key: "_id", Value: -1}}}},
		// เผื่อหน้า +1 เพื่อดูว่ามี next ไหม
		bson.D{{Key: "$limit", Value: limit + 1}},
	)

	cctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cur, e := postsColl.Aggregate(cctx, pipe, options.Aggregate())
	if e != nil {
		return nil, nil, 0, e
	}
	defer cur.Close(cctx)

	var all []bson.M
	if e := cur.All(cctx, &all); e != nil {
		return nil, nil, 0, e
	}

	// 4) ตัดหน้า + next cursor
	if int64(len(all)) > limit {
		items = all[:limit]
		last := items[len(items)-1]

		// parse created_at
		var tm time.Time
		switch v := last["created_at"].(type) {
		case time.Time:
			tm = v.UTC()
		case bson.DateTime:
			tm = time.UnixMilli(int64(v)).UTC()
		case primitive.DateTime:
			tm = v.Time().UTC()
		case string:
			if t2, perr := time.Parse(time.RFC3339Nano, v); perr == nil {
				tm = t2.UTC()
			} else if t2, perr := time.Parse(time.RFC3339, v); perr == nil {
				tm = t2.UTC()
			} else {
				return nil, nil, 0, fmt.Errorf("invalid created_at string: %q", v)
			}
		default:
			return nil, nil, 0, fmt.Errorf("unknown created_at type: %T", v)
		}

		// parse _id
		var lastID primitive.ObjectID
		switch v := last["_id"].(type) {
		case primitive.ObjectID:
			lastID = v
		case bson.ObjectID:
			if oid, perr := primitive.ObjectIDFromHex(v.Hex()); perr == nil {
				lastID = oid
			} else {
				return nil, nil, 0, fmt.Errorf("cannot convert bson.ObjectID: %v", perr)
			}
		case string:
			if oid, perr := primitive.ObjectIDFromHex(v); perr == nil {
				lastID = oid
			} else {
				return nil, nil, 0, fmt.Errorf("invalid _id hex in page item: %q", v)
			}
		default:
			return nil, nil, 0, fmt.Errorf("unknown _id type: %T", v)
		}

		s := cursor.EncodePostCursor(tm, lastID)
		next = &s
	} else {
		items = all
		next = nil
	}

	return
}
