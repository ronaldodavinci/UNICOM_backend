package services

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"like_workspace/dto"
	"like_workspace/internal/repository"
	"like_workspace/model"
)

// public entry (handler เรียกฟังก์ชันนี้ฟังก์ชันเดียว)
func Like(ctx context.Context, client *mongo.Client, body dto.LikeRequestDTO) (int, any) {
	if err := validateLikeRequest(body); err != nil {
		return fiberStatusBadRequest(), dto.ErrorResponse{Message: err.Error()}
	}
	userID, targetID, err := parseObjectIDs(body)
	if err != nil {
		return fiberStatusBadRequest(), dto.ErrorResponse{Message: err.Error()}
	}

	db := client.Database("lll_workspace")
	likesCol := db.Collection("like")

	updateCol, likeDoc, targetFilter := buildLikeDocAndTarget(db, body, userID, targetID, time.Now().UTC())

	dup, err := repository.InsertLike(ctx, likesCol, likeDoc)
	if err != nil { // error อื่นๆ
		return fiberStatusInternalError(), dto.ErrorResponse{Message: err.Error()}
	}
	if dup { // duplicate (เคยกดไลค์แล้ว)
		return fiberStatusOK(), bson.M{
			"message":     "already-liked",
			"target_id":   body.TargetID,
			"target_type": body.TargetType,
			"is_liked":    true,
		}
	}

	_ = repository.IncLikeCount(ctx, updateCol, targetFilter)

	return fiberStatusCreated(), bson.M{
		"message":     "liked",
		"target_id":   body.TargetID,
		"target_type": body.TargetType,
		"is_liked":    true,
	}
}

// helper

func validateLikeRequest(body dto.LikeRequestDTO) error {
	if body.UserID == "" || body.TargetType == "" || body.TargetID == "" {
		return errors.New("userId, targetType, targetId are required")
	}
	if body.TargetType != "post" && body.TargetType != "comment" {
		return errors.New("targetType must be 'post' or 'comment'")
	}
	return nil
}

// แปลงเป็น ObjectID พร้อมตรวจสอบความถูกต้อง
func parseObjectIDs(body dto.LikeRequestDTO) (userID bson.ObjectID, targetID bson.ObjectID, err error) {
	uid, uerr := bson.ObjectIDFromHex(body.UserID)   // เดิม
	tid, terr := bson.ObjectIDFromHex(body.TargetID) // เดิม
	if uerr != nil {
		return bson.ObjectID{}, bson.ObjectID{}, errors.New("invalid userId")
	}
	if terr != nil {
		return bson.ObjectID{}, bson.ObjectID{}, errors.New("invalid targetId")
	}
	return uid, tid, nil
}

// สร้าง likeDoc, กำหนด target collection(collection ที่เราจะ insert -> post/comment), ทำ filter เอาไปใช้ใน update (inc +1 like_count)
// คืนค่า: collection เป้าหมาย + likeDoc + filter เดิม
func buildLikeDocAndTarget(db *mongo.Database, body dto.LikeRequestDTO, userID bson.ObjectID, targetID bson.ObjectID, now time.Time) (*mongo.Collection, model.Like, bson.M) {
	var updateCol *mongo.Collection
	var likeDoc model.Like

	if body.TargetType == "post" {
		updateCol = db.Collection("posts")
		likeDoc = model.Like{
			UserID:    userID, // ใช้ชนิดเดิมของคุณ
			PostID:    &targetID,
			CommentID: nil,
			CreatedAt: now,
		}
	} else { // "comment"
		updateCol = db.Collection("comments")
		likeDoc = model.Like{
			UserID:    userID,
			PostID:    nil,
			CommentID: &targetID,
			CreatedAt: now,
		}
	}
	targetFilter := bson.M{"_id": targetID}
	return updateCol, likeDoc, targetFilter
}

// mapping สถานะ HTTP (แยกไว้ให้ชัด ไม่แก้ตัวเลขเดิม)
func fiberStatusOK() int            { return 200 }
func fiberStatusCreated() int       { return 201 }
func fiberStatusBadRequest() int    { return 400 }
func fiberStatusInternalError() int { return 500 }
