package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)
type NotiType string

type Ref struct {
	Entity string             `bson:"entity" json:"entity"` // "event" | "qa"
	ID     bson.ObjectID `bson:"id"     json:"id"`     // eventId หรือ answerId
}

type Notification struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    bson.ObjectID `bson:"user_id"       json:"userId"`
	Type      NotiType           `bson:"type"          json:"type"`
	Title     string             `bson:"title"         json:"title"`
	Body      string             `bson:"body"          json:"body"`
	Ref       Ref                `bson:"ref"           json:"ref"`
	CreatedAt time.Time          `bson:"created_at"    json:"createdAt"`
	Read bool			   `bson:"read"          json:"read"`
}

// ---- 1) ฟังก์ชันกลางสร้าง title/body จาก type + params ----
type NotiParams struct {
	EventTitle string // ใช้กับหลายเคส
	EventID	bson.ObjectID
	StartTime *time.Time // ใช้กับ event reminder
	// เติม field อื่นได้ถ้าต้องใช้ในอนาคต
}