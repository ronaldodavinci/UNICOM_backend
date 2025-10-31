package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	m "main-webbase/internal/models"
)

const (
	NotiAuditionApproved m.NotiType = "AUDITION_APPROVED"
	NotiEventUpdated     m.NotiType = "EVENT_UPDATED"
	NotiEventDeleted     m.NotiType = "EVENT_DELETED"
	NotiEventReminder    m.NotiType = "EVENT_REMINDER"
	NotiQAAnswered       m.NotiType = "QA_ANSWERED"
	NotiQAQuestion       m.NotiType = "QA_QUESTION"
)

func BuildTitleBody(t m.NotiType, p m.NotiParams) (title, body string, err error) {
	switch t {
	// delete
	case NotiEventDeleted:
		if p.EventTitle == "" {
			return "", "", errors.New("missing EventTitle")
		}
		return fmt.Sprintf("%s is deleted", p.EventTitle),
			fmt.Sprintf("%s is deleted.", p.EventTitle), nil

	case NotiEventUpdated:
		if p.EventTitle == "" {
			return "", "", errors.New("missing EventTitle")
		}
		return fmt.Sprintf("%s has been updated", p.EventTitle),
			fmt.Sprintf("%s has been updated. Please check the details.", p.EventTitle), nil

	case NotiAuditionApproved:
		if p.EventTitle == "" {
			return "", "", errors.New("missing EventTitle")
		}
		return "Audition approved 🎉",
			fmt.Sprintf("You are approved for %s. Please check the details.", p.EventTitle), nil

	case NotiEventReminder:
		if p.EventTitle == "" || p.StartTime == nil {
			return "", "", errors.New("missing EventTitle/StartTime")
		}
		return "Event reminder",
			fmt.Sprintf("%s is coming up. Please check the details.", p.EventTitle), nil

	case NotiQAAnswered:
		if p.EventTitle == "" {
			return "", "", errors.New("missing EventTitle")
		}
		return "Your question has a new answer",
			fmt.Sprintf("Your question on %s has been answered.", p.EventTitle), nil
	case NotiQAQuestion:
		if p.EventTitle == "" {
			return "", "", errors.New("missing EventTitle")
		}
		return "Your event has a new question",
			fmt.Sprintf("A new question was posted on %s event.", p.EventTitle), nil
	}
	return "", "", fmt.Errorf("unknown noti type: %s", t)
}

// ตัวช่วยยิง noti (One/Many)

// สร้าง notification ให้ user คนเดียว
func NotifyOne(ctx context.Context, col *mongo.Collection,
	userID bson.ObjectID, typ m.NotiType, ref m.Ref, p m.NotiParams) error {

	title, body, err := BuildTitleBody(typ, p)
	if err != nil {
		return err
	}
	_, err = col.InsertOne(ctx, bson.M{
		"user_id":    userID,
		"type":       typ,
		"title":      title,
		"body":       body,
		"ref":        ref,
		"read":       false,
		"created_at": time.Now().UTC(),
	})
	return err
}

// สร้าง notification ให้หลาย user พร้อมกัน -->  delete, update, reminder
func NotifyMany(ctx context.Context, col *mongo.Collection,
	userIDs []bson.ObjectID, typ m.NotiType, ref m.Ref, p m.NotiParams) error {

	if len(userIDs) == 0 {
		return nil
	}
	title, body, err := BuildTitleBody(typ, p)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	writes := make([]mongo.WriteModel, 0, len(userIDs))
	for _, uid := range userIDs {
		if uid.IsZero() {
			return fmt.Errorf("notifyMany: found zero userID in payload")
		}
		writes = append(writes, &mongo.InsertOneModel{Document: bson.M{
			"user_id":    uid,
			"type":       typ,
			"title":      title,
			"body":       body,
			"ref":        ref,
			"created_at": now,
			"read":       false,
		}})
	}
	_, err = col.BulkWrite(ctx, writes, options.BulkWrite().SetOrdered(false))
	return err
}

// reminder
func RunEventReminder(ctx context.Context, db *mongo.Database, loc *time.Location, horizonDays int) error {

	colSched := db.Collection("event_schedules")
	colRegs := db.Collection("event_participant")
	colNoti := db.Collection("notification")

	nowLocal := time.Now().In(loc)
	nowUTC := nowLocal.In(time.UTC)
	endUTC := nowUTC.Add(time.Duration(horizonDays) * 24 * time.Hour)

	const reminderDaysBefore = 7
	const catchUpDays = 14

	advanceDuration := time.Duration(reminderDaysBefore) * 24 * time.Hour
	catchUpDuration := time.Duration(catchUpDays) * 24 * time.Hour

	lowerBound := nowUTC.Add(-catchUpDuration + advanceDuration)

	// หา first_day ต่อ event_id แล้วคัดเฉพาะที่ยังไม่เริ่มและอยู่ในช่วง [now, now+horizon]
	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$event_id"},
			{Key: "first_day", Value: bson.D{{Key: "$min", Value: "$date"}}},
		}}},
		{{Key: "$match", Value: bson.D{
			{Key: "first_day", Value: bson.D{
				{Key: "$gte", Value: lowerBound},
				{Key: "$lte", Value: endUTC},
			}},
		}}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "events"},
			{Key: "localField", Value: "_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "event"},
		}}},
		{{Key: "$unwind", Value: "$event"}},
		{{Key: "$match", Value: bson.D{{Key: "event.status", Value: "active"}}}},
	}

	cur, err := colSched.Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	type rowT struct {
		EventID  bson.ObjectID `bson:"_id"`
		FirstDay time.Time     `bson:"first_day"`
		Event    struct {
			ID     bson.ObjectID `bson:"_id"`
			Title  string        `bson:"topic"`
			Status string        `bson:"status"`
		} `bson:"event"`
	}

	var writes []mongo.WriteModel
	now := time.Now().UTC()

	for cur.Next(ctx) {
		var row rowT
		if err := cur.Decode(&row); err != nil {
			return err
		}

		if row.FirstDay.Before(nowUTC) {
			continue
		}

		reminderAt := row.FirstDay.Add(-advanceDuration)
		if reminderAt.After(nowUTC) {
			// Not yet time to send reminder
			continue
		}
		if nowUTC.Sub(reminderAt) > catchUpDuration {
			// Missed reminder outside catch-up window
			continue
		}

		// ดึงผู้ลงทะเบียน
		regCur, err := colRegs.Find(ctx, bson.M{
			"event_id": row.EventID,
			"status":   "accept",
			"role":     "participant",
		}, options.Find().SetProjection(bson.M{"user_id": 1}))
		if err != nil {
			return err
		}

		var userIDs []bson.ObjectID
		for regCur.Next(ctx) {
			var r struct {
				UserID bson.ObjectID `bson:"user_id"`
			}
			if err := regCur.Decode(&r); err != nil {
				regCur.Close(ctx)
				return err
			}
			userIDs = append(userIDs, r.UserID)
		}
		regCur.Close(ctx)
		if len(userIDs) == 0 {
			continue
		}

		// เตรียมข้อความ
		startLocal := row.FirstDay.In(loc)
		p := m.NotiParams{EventTitle: row.Event.Title, EventID: row.EventID, StartTime: &startLocal}
		title, body, err := BuildTitleBody(NotiEventReminder, p)
		if err != nil {
			continue
		}

		// Upsert กันซ้ำต่อ user/event/first_day (idempotent)
		for _, uid := range userIDs {
			writes = append(writes, &mongo.ReplaceOneModel{
				Filter: bson.M{
					"user_id":        uid,
					"type":           string(NotiEventReminder),
					"ref.event_id":   row.EventID,
					"meta.first_day": row.FirstDay,
				},
				Replacement: bson.M{
					"user_id": uid,
					"type":    string(NotiEventReminder),
					"title":   title,
					"body":    body,
					"ref":     bson.M{"event_id": row.EventID},
					"meta": bson.M{
						"first_day":            row.FirstDay,
						"reminder_days_before": reminderDaysBefore,
					},
					"created_at": now,
					"read":       false,
				},
				Upsert: boolPtr(true),
			})
		}
	}

	if len(writes) > 0 {
		_, err = colNoti.BulkWrite(ctx, writes, options.BulkWrite().SetOrdered(false))
		if err != nil {
			return err
		}
	}
	return nil
}

func boolPtr(b bool) *bool { return &b }
