package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"like_workspace/dto"
	"like_workspace/model"
	repo "like_workspace/internal/repository"
	u "like_workspace/internal/utils"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var ErrUserNotFound = errors.New("user not found")
var ErrOrgNodeNotFound = errors.New("org node not found")
var ErrPositionNotFound = errors.New("position not found")


func CreatePostWithMeta(client *mongo.Client, UserID bson.ObjectID, body dto.CreatePostDTO, ctx context.Context) (dto.PostResponse, error) {
	db := client.Database("lll_workspace")
	now := time.Now().UTC()

	var resp dto.PostResponse
	postsCol := db.Collection("posts")

	// 0) เตรียม RolePathID / PositionID จาก DTO (lookup ด้วย org_path, position_key)
	rolePathID, err := repo.ResolveOrgNodeIDByPath(db, body.PostAs.OrgPath, ctx)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return resp, ErrOrgNodeNotFound
		}
		return resp, err
	}

	positionID, err := repo.ResolvePositionIDByKey(db, body.PostAs.PositionKey, ctx)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return resp, ErrPositionNotFound
		}
		return resp, err
	}

	// 0.1) เตรียม tags จาก PostText
	tagsSlice := u.ExtractHashtags(body.PostText)

	// 1) Insert post
	post := model.Post{
		UserID:       UserID,
		RolePathID:   rolePathID,   // เปลี่ยนจาก RolePath(string) → ObjectID
		PositionID:   positionID,   // เปลี่ยนจาก Position(string) → ObjectID
		Hashtag:         tagsSlice,      // เก็บ string (เช่น "smo,eng,ku66")
		Tags: body.PostAs.Tag,
		PostText:     body.PostText,
		CreatedAt:    now,
		UpdatedAt:    now,
		LikeCount:    0,
		CommentCount: 0,
	}

	res, err := postsCol.InsertOne(ctx, post)
	if err != nil {
		return resp, err
	}
	fmt.Println("🆗 post created with ID:", res.InsertedID)
	post.ID = res.InsertedID.(bson.ObjectID)

	// helper: rollback ทุกอย่างที่อาจสร้างไปแล้ว (best-effort)
	rollback := func() {
		_, _ = postsCol.DeleteOne(ctx, bson.M{"_id": post.ID})
		_, _ = db.Collection("post_categories").DeleteMany(ctx, bson.M{"post_id": post.ID})
		_, _ = db.Collection("post_rolevisible").DeleteMany(ctx, bson.M{"post_id": post.ID})
		_, _ = db.Collection("post_hashtag").DeleteMany(ctx, bson.M{"post_id": post.ID})
	}

	// 2) hashtags (non-critical; ลงทั้งตาราง post_hashtag และเก็บ string ใน post.Tags แล้ว)
	if err := repo.InsertHashtags(db, post, body.PostText, ctx); err != nil {
		fmt.Println("⚠️ hashtag insert failed:", err)
	}

	// 3) categories (critical)
	if len(body.CategoryIDs) > 0 {
		if err := repo.InsertCategories(db, post.ID, body.CategoryIDs, ctx); err != nil {
			rollback()
			return resp, err
		}
	}

	// 4) role visibility (critical): ACCESS=private → บันทึกลง post_rolevisible โดยแปลง org_path → node_id (ObjectID)
	if body.Visibility.Access == "private" {
		if err := repo.InsertRoleVisibility(db, post.ID, body.Visibility, ctx); err != nil {
			rollback()
			return resp, err
		}
	}


	// 5) ดึง user info (critical)
	userInfo, err := repo.FindUserInfo(db, UserID, ctx)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			rollback()
			return resp, ErrUserNotFound
		}
		rollback()
		return resp, err
	}

	// 6) ประกอบ response (ส่ง string id กลับตาม requirement)
	resp = dto.PostResponse{
		UserID:        UserID.Hex(),
		Name:          userInfo.FirstName, // แก้เป็น display name ที่ต้องการได้
		Username:      userInfo.Username,
		PostText:      post.PostText,
		LikeCount:     post.LikeCount,
		CommentCount:  post.CommentCount,
		LikedBy:       []string{},
		PostAs:        body.PostAs,
		CategoryIDs:   body.CategoryIDs,       // ถ้าในระบบเป็น ObjectID ให้ map เป็น hex ก่อน
		Visibility:    body.Visibility,
		OrgOfContent:  body.PostAs.OrgPath,    // ส่ง org_path ให้ FE
		CreatedAt:     post.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     post.UpdatedAt.Format(time.RFC3339),
		Status:        "active",
	}

	return resp, nil
}
