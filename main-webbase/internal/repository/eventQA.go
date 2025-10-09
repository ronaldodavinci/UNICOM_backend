package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// FindEventOrganizerID คืน organizer_id ของอีเวนต์ (หรือ mongo.ErrNoDocuments ถ้าไม่พบ)
func FindEventOrganizerID(col *mongo.Collection, eventID bson.ObjectID, ctx context.Context) (bson.ObjectID, error) {
	var doc struct {
		OrganizerID bson.ObjectID `bson:"node_id"`
	}
	err := col.FindOne(ctx,
				bson.M{"_id": eventID},
				options.FindOne().SetProjection(bson.M{"node_id": 1}),
				).Decode(&doc)
	if err != nil {
		return bson.ObjectID{}, err // รวม mongo.ErrNoDocuments ให้ handler ตัดสินใจต่อ
	}
	return doc.OrganizerID, nil
}
