package services

import (
	"context"
	"main-webbase/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Service struct {
	col *mongo.Collection
}

func NewService(col *mongo.Collection) *Service {
	return &Service{col: col}
}

func (s *Service) List(ctx context.Context, opt models.QueryOptions) ([]models.Post, *bson.ObjectID, error) {
	if opt.Limit <= 0 || opt.Limit > 100 {
		opt.Limit = 20
	}

	and := make([]bson.M, 0, 6)

	if !opt.UntilID.IsZero() {
		and = append(and, bson.M{"_id": bson.M{"$lt": opt.UntilID}})
	} else if !opt.SinceID.IsZero() {
		and = append(and, bson.M{"_id": bson.M{"$gt": opt.SinceID}})
	}

	if len(opt.AuthorIDs) > 0 {
		and = append(and, bson.M{"user_id": bson.M{"$in": opt.AuthorIDs}})
	}
	if opt.TextSearch != "" {
		and = append(and, bson.M{"post_text": bson.M{"$regex": opt.TextSearch, "$options": "i"}})
	}
	if len(opt.Categories) > 0 {
		and = append(and, bson.M{"categories": bson.M{"$in": opt.Categories}})
	}
	if len(opt.Tags) > 0 {
		and = append(and, bson.M{"tags": bson.M{"$in": opt.Tags}})
	}

	// owner เห็นเสมอ
	ownerOr := []bson.M{}
	if !opt.ViewerID.IsZero() {
		ownerOr = append(ownerOr, bson.M{"user_id": opt.ViewerID})
	}

	// หมายเหตุสำคัญ:
	// ถ้าคุณต้องใช้เงื่อนไข public/private จาก post_role_visibility (collection แยก)
	// การใช้ .Find() ตรงๆ ที่ posts จะไม่พอ — ต้องใช้ aggregate + $lookup post_role_visibility
	// ทาง services.Find() แบบนี้จะใช้ได้ก็ต่อเมื่อคุณเก็บ flag is_public หรือ flatten vis roles ไว้ใน posts อยู่แล้ว

	// ตัวอย่างง่ายๆ: ถ้าคุณมีฟิลด์ is_public ใน posts
	visOr := []bson.M{
		{"is_public": true},
	}
	if len(opt.AllowedNodeIDs) > 0 {
		// ถ้าใน posts ไม่มี vis ฝังไว้ ต้องย้ายไป aggregate ตามที่เราคุยก่อนหน้า
		// ที่นี่สมมุติว่ามี array vis_node_ids ใน posts
		visOr = append(visOr, bson.M{"vis_node_ids": bson.M{"$in": opt.AllowedNodeIDs}})
	}
	visOr = append(visOr, ownerOr...)

	and = append(and, bson.M{"$or": visOr})

	filter := bson.M{}
	if len(and) == 1 {
		filter = and[0]
	} else {
		filter = bson.M{"$and": and}
	}

	findOpt := options.Find().
		SetSort(bson.M{"_id": -1}).
		SetLimit(opt.Limit + 1)

	cur, err := s.col.Find(ctx, filter, findOpt)
	if err != nil {
		return nil, nil, err
	}
	defer cur.Close(ctx)

	var items []models.Post
	if err := cur.All(ctx, &items); err != nil {
		return nil, nil, err
	}

	var next *bson.ObjectID
	if int64(len(items)) == opt.Limit+1 {
		last := items[len(items)-1].ID
		items = items[:len(items)-1]
		next = &last
	}
	return items, next, nil
}
