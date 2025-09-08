// internal/repository/posts_repository.go
package repository

import (
	"context"
	"time"
	"fmt"

	"like_workspace/internal/cursor"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongo "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Post = bson.M

// ListPostsInAnyCategoryNewestFirst
// ดึง "ทุกโพสต์ที่มีอยู่ใน post_categories อย่างน้อย 1 รายการ" (ไม่กรองตามหมวด)
// เรียงใหม่ก่อนเก่า และทำ cursor pagination (created_at, _id)
func ListPostsInAnyCategoryNewestFirst(
	ctx context.Context,
	client *mongo.Client,
	cursorStr string,
	limit int64,
) (items []bson.M, next *string, total int64, err error) {

	db := client.Database("lll_workspace")
	pcColl := db.Collection("post_categories")

	// 1) total (นับเฉพาะโพสต์ที่มีอยู่จริง)
	countPipe := mongo.Pipeline{
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "posts"},
			{Key: "localField", Value: "post_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "post"},
		}}},
		{{Key: "$unwind", Value: "$post"}},
		{{Key: "$group", Value: bson.D{{Key: "_id", Value: "$post._id"}}}},
		{{Key: "$count", Value: "n"}},
	}
	{
		cctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		cur, e := pcColl.Aggregate(cctx, countPipe)
		if e != nil {
			err = e
			return
		}
		var row []struct{ N int64 `bson:"n"` }
		e = cur.All(cctx, &row)
		if e != nil {
			err = e
			return
		}
		if len(row) > 0 {
			total = row[0].N
		}
	}

	// 2) cursor match
	var cursorMatch bson.D
	if cursorStr != "" {
		t, oid, e := cursor.DecodePostCursor(cursorStr)
		if e != nil {
			err = e
			return
		}
		cursorMatch = bson.D{{
			Key: "$or",
			Value: bson.A{
				bson.D{{Key: "post.created_at", Value: bson.D{{Key: "$lt", Value: t}}}},
				bson.D{{Key: "post.created_at", Value: t}, {Key: "post._id", Value: bson.D{{Key: "$lt", Value: oid}}}},
			},
		}}
	}

	// 3) pipeline หลัก
	pipe := mongo.Pipeline{
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "posts"},
			{Key: "localField", Value: "post_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "post"},
		}}},
		{{Key: "$unwind", Value: "$post"}},
	}
	if len(cursorMatch) > 0 {
		pipe = append(pipe, bson.D{{Key: "$match", Value: cursorMatch}})
	}
	pipe = append(pipe,
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$post._id"},
			{Key: "post", Value: bson.D{{Key: "$first", Value: "$post"}}},
			{Key: "matched_category_ids", Value: bson.D{{Key: "$addToSet", Value: "$category_id"}}},
		}}},
		bson.D{{Key: "$replaceRoot", Value: bson.D{
			{Key: "newRoot", Value: bson.D{
				{Key: "$mergeObjects", Value: bson.A{
					"$post",
					bson.D{{Key: "matched_category_ids", Value: "$matched_category_ids"}},
				}},
			}},
		}}},
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "created_at", Value: bson.D{{Key: "$toDate", Value: "$created_at"}}},
			{Key: "like_count", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$like_count", 0}}}},
			{Key: "comment_count", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$comment_count", 0}}}},
		}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "created_at", Value: -1}, {Key: "_id", Value: -1}}}},
		bson.D{{Key: "$limit", Value: limit + 1}},
	)

	cctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cur, e := pcColl.Aggregate(cctx, pipe, options.Aggregate())
	if e != nil {
		err = e
		return
	}
	defer cur.Close(cctx)

	var all []bson.M
	e = cur.All(cctx, &all)
	if e != nil {
		err = e
		return
	}

	// 4) ตัดหน้า + next cursor
	if int64(len(all)) > limit {
		items = all[:limit]
		last := items[len(items)-1]

		// created_at: รองรับหลายชนิด (bson.DateTime/time.Time/string/primitive.DateTime)
		var tm time.Time
		switch v := last["created_at"].(type) {
		case time.Time:
			tm = v.UTC()
		case bson.DateTime:
			tm = time.UnixMilli(int64(v)).UTC()
		case primitive.DateTime:
			tm = v.Time().UTC()
		case string:
			if t, perr := time.Parse(time.RFC3339Nano, v); perr == nil {
				tm = t.UTC()
			} else if t, perr := time.Parse(time.RFC3339, v); perr == nil {
				tm = t.UTC()
			} else {
				err = fmt.Errorf("invalid created_at string: %q", v)
				return
			}
		default:
			err = fmt.Errorf("unknown created_at type: %T", v)
			return
		}

		// _id: รองรับ primitive.ObjectID / bson.ObjectID / string(hex)
		var lastID primitive.ObjectID
		switch v := last["_id"].(type) {
		case primitive.ObjectID:
			lastID = v
		case bson.ObjectID:
			if oid, perr := primitive.ObjectIDFromHex(v.Hex()); perr == nil {
				lastID = oid
			} else {
				err = fmt.Errorf("cannot convert bson.ObjectID: %v", perr)
				return
			}
		case string:
			if oid, perr := primitive.ObjectIDFromHex(v); perr == nil {
				lastID = oid
			} else {
				err = fmt.Errorf("invalid _id hex in page item: %q", v)
				return
			}
		default:
			err = fmt.Errorf("unknown _id type: %T", v)
			return
		}

		s := cursor.EncodePostCursor(tm, lastID)
		next = &s
	} else {
		items = all
		next = nil
	}

	return
}

// ListPostsByCategoryNewestFirst
// ดึงโพสต์ที่อยู่ใน "ชุดหมวดหมู่ที่กำหนด" อย่างน้อยหนึ่งหมวด
// ใหม่ก่อนเก่า + cursor (created_at, _id) + total แบบ lookup จริง
func ListPostsByCategoryNewestFirst(
	ctx context.Context,
	client *mongo.Client,
	categoryIDs []bson.ObjectID,
	cursorStr string,
	limit int64,
) (items []bson.M, next *string, total int64, err error) {

	db := client.Database("lll_workspace")
	pcColl := db.Collection("post_categories")

	if len(categoryIDs) == 0 {
		err = fmt.Errorf("empty categoryIDs")
		return
	}

	// 1) total (distinct post._id ที่มีอยู่จริง + อยู่ในหมวดที่กำหนด)
	countPipe := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{{Key: "category_id", Value: bson.D{{Key: "$in", Value: categoryIDs}}}}}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "posts"},
			{Key: "localField", Value: "post_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "post"},
		}}},
		{{Key: "$unwind", Value: "$post"}},
		{{Key: "$group", Value: bson.D{{Key: "_id", Value: "$post._id"}}}},
		{{Key: "$count", Value: "n"}},
	}
	{
		cctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		cur, e := pcColl.Aggregate(cctx, countPipe)
		if e != nil {
			err = e
			return
		}
		var row []struct{ N int64 `bson:"n"` }
		e = cur.All(cctx, &row)
		if e != nil {
			err = e
			return
		}
		if len(row) > 0 {
			total = row[0].N
		}
	}

	// 2) cursor match
	var cursorMatch bson.D
	if cursorStr != "" {
		t, oid, e := cursor.DecodePostCursor(cursorStr)
		if e != nil {
			err = e
			return
		}
		cursorMatch = bson.D{{
			Key: "$or",
			Value: bson.A{
				bson.D{{Key: "post.created_at", Value: bson.D{{Key: "$lt", Value: t}}}},
				bson.D{{Key: "post.created_at", Value: t}, {Key: "post._id", Value: bson.D{{Key: "$lt", Value: oid}}}},
			},
		}}
	}

	// 3) pipeline หลัก
	pipe := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{{Key: "category_id", Value: bson.D{{Key: "$in", Value: categoryIDs}}}}}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "posts"},
			{Key: "localField", Value: "post_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "post"},
		}}},
		{{Key: "$unwind", Value: "$post"}},
	}
	if len(cursorMatch) > 0 {
		pipe = append(pipe, bson.D{{Key: "$match", Value: cursorMatch}})
	}
	pipe = append(pipe,
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$post._id"},
			{Key: "post", Value: bson.D{{Key: "$first", Value: "$post"}}},
			{Key: "matched_category_ids", Value: bson.D{{Key: "$addToSet", Value: "$category_id"}}},
		}}},
		bson.D{{Key: "$replaceRoot", Value: bson.D{
			{Key: "newRoot", Value: bson.D{
				{Key: "$mergeObjects", Value: bson.A{
					"$post",
					bson.D{{Key: "matched_category_ids", Value: "$matched_category_ids"}},
				}},
			}},
		}}},
		// บังคับชนิดเวลา + ค่าเริ่มต้น count
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "created_at", Value: bson.D{{Key: "$toDate", Value: "$created_at"}}},
			{Key: "like_count", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$like_count", 0}}}},
			{Key: "comment_count", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$comment_count", 0}}}},
		}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "created_at", Value: -1}, {Key: "_id", Value: -1}}}},
		bson.D{{Key: "$limit", Value: limit + 1}},
	)

	cctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cur, e := pcColl.Aggregate(cctx, pipe, options.Aggregate())
	if e != nil {
		err = e
		return
	}
	defer cur.Close(cctx)

	var all []bson.M
	e = cur.All(cctx, &all)
	if e != nil {
		err = e
		return
	}

	// 4) ตัดหน้า + next cursor (รองรับหลายชนิดเวลา/ไอดี)
	if int64(len(all)) > limit {
		items = all[:limit]
		last := items[len(items)-1]

		var tm time.Time
		switch v := last["created_at"].(type) {
		case time.Time:
			tm = v.UTC()
		case bson.DateTime:
			tm = time.UnixMilli(int64(v)).UTC()
		case primitive.DateTime:
			tm = v.Time().UTC()
		case string:
			if t, perr := time.Parse(time.RFC3339Nano, v); perr == nil {
				tm = t.UTC()
			} else if t, perr := time.Parse(time.RFC3339, v); perr == nil {
				tm = t.UTC()
			} else {
				err = fmt.Errorf("invalid created_at string: %q", v)
				return
			}
		default:
			err = fmt.Errorf("unknown created_at type: %T", v)
			return
		}

		var lastID primitive.ObjectID
		switch v := last["_id"].(type) {
		case primitive.ObjectID:
			lastID = v
		case bson.ObjectID:
			if oid, perr := primitive.ObjectIDFromHex(v.Hex()); perr == nil {
				lastID = oid
			} else {
				err = fmt.Errorf("cannot convert bson.ObjectID: %v", perr)
				return
			}
		case string:
			if oid, perr := primitive.ObjectIDFromHex(v); perr == nil {
				lastID = oid
			} else {
				err = fmt.Errorf("invalid _id hex in page item: %q", v)
				return
			}
		default:
			err = fmt.Errorf("unknown _id type: %T", v)
			return
		}

		s := cursor.EncodePostCursor(tm, lastID)
		next = &s
	} else {
		items = all
		next = nil
	}

	return
}
