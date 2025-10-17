package services

import (
	"context"
	"errors"
	"fmt"
	"main-webbase/dto"
	"main-webbase/internal/models"
	repo "main-webbase/internal/repository"
	u "main-webbase/internal/utils"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var ErrUserNotFound = errors.New("user not found")
var ErrOrgNodeNotFound = errors.New("org node not found")
var ErrPositionNotFound = errors.New("position not found")
var ErrUserIDInvalid = errors.New("userID invalid")

func CreatePostWithMeta(client *mongo.Client, UserID string, body dto.CreatePostDTO, ctx context.Context) (dto.PostResponse, error) {
	db := client.Database("unicom")
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

	UserIDs, err := bson.ObjectIDFromHex(UserID)
	if err != nil {
		return resp, ErrUserIDInvalid
	}
	// 1) Insert post
	post := models.Post{
		UserID:       UserIDs,
		RolePathID:   rolePathID, // เปลี่ยนจาก RolePath(string) → ObjectID
		PositionID:   positionID, // เปลี่ยนจาก Position(string) → ObjectID
		Hashtag:      tagsSlice,  // เก็บ string (เช่น "smo,eng,ku66")
		Tags:         body.PostAs.Tag,
		PostText:     body.PostText,
		Media:        body.Media,
		CreatedAt:    now,
		UpdatedAt:    now,
		LikeCount:    0,
		CommentCount: 0,
		Status:       "active",
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
		_, _ = db.Collection("post_role_visible").DeleteMany(ctx, bson.M{"post_id": post.ID})
		_, _ = db.Collection("hashtags").DeleteMany(ctx, bson.M{"post_id": post.ID})
	}

	// 2) hashtags (non-critical; ลงทั้งตาราง post_hashtag และเก็บ string ใน post.Tags แล้ว)
	if err := repo.RebuildHashtags(db, post, body.PostText, ctx); err != nil {
		fmt.Println("⚠️ hashtag insert failed:", err)
	}

	// 3) categories (critical)
	if len(body.CategoryIDs) > 0 {
		if err := repo.ReplaceCategories(db, post.ID, body.CategoryIDs, ctx); err != nil {
			rollback()
			return resp, err
		}
	}

	// 4) role visibility (critical): ACCESS=private → บันทึกลง post_rolevisible โดยแปลง org_path → node_id (ObjectID)
	if body.Visibility.Access == "private" {
		if err := repo.ReplaceRoleVisibility(db, post.ID, body.Visibility, ctx); err != nil {
			rollback()
			return resp, err
		}
	}

	// 5) ดึง user info (critical)
	colUsers := db.Collection("users")
	userInfo, err := repo.FindUserInfo(colUsers, UserIDs, ctx)
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
		UserID:       UserID,
		Name:         userInfo.FirstName, // แก้เป็น display name ที่ต้องการได้
		Username:     userInfo.Username,
		PostText:     post.PostText,
		Hashtag:      post.Hashtag,
		LikeCount:    post.LikeCount,
		CommentCount: post.CommentCount,
		// LikedBy:      []string{},
		PostAs:       body.PostAs,
		CategoryIDs:  body.CategoryIDs, // ถ้าในระบบเป็น ObjectID ให้ map เป็น hex ก่อน
		Visibility:   body.Visibility,
		OrgOfContent: body.PostAs.OrgPath, // ส่ง org_path ให้ FE
		CreatedAt:    post.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    post.UpdatedAt.Format(time.RFC3339),
		Status:       "active",
	}

	return resp, nil
}

func GetPostDetail(ctx context.Context, db *mongo.Database, loginUserID bson.ObjectID, postID bson.ObjectID) (dto.PostResponse, error) {
	var out dto.PostResponse

	colPosts := db.Collection("posts")
	colUsers := db.Collection("users")
	colPositions := db.Collection("positions")
	colOrgNodes := db.Collection("org_units")
	colPostRoleVis := db.Collection("post_role_visibility")
	colCats := db.Collection("post_categories")
	colLikes := db.Collection("like")

	// 1) post
	post, err := repo.FindPostByID(colPosts, postID, ctx)
	if err != nil {
		return out, fmt.Errorf("post not found or fetch error: %w", err)
	}

	if post.Status != "active" {
		return out, fmt.Errorf("post is not active")
	}

	// 2) user
	user, err := repo.FindUserInfo(colUsers, post.UserID, ctx)
	if err != nil {
		return out, fmt.Errorf("fetch user: %w", err)
	}
	fullName := user.FirstName
	if user.LastName != "" {
		fullName = user.FirstName + " " + user.LastName
	}

	// 3) position
	posName := "Unknown Position"
	if !post.PositionID.IsZero() {
		if key, err := repo.FindPositionName(colPositions, post.PositionID, ctx); err == nil && key != "" {
			posName = key
		}
	}

	// 4) org (short_name/path จาก repo.FindOrgNode)
	orgPath := ""
	if !post.RolePathID.IsZero() {
		if n, err := repo.FindOrgNode(colOrgNodes, post.RolePathID, ctx); err == nil {
			orgPath = n
		}
	}

	// 5) visibility
	vis, err := repo.FindVisibilityPaths(colPostRoleVis, colOrgNodes, post.ID, ctx)
	if err != nil {
		return out, fmt.Errorf("fetch visibility: %w", err)
	}

	// 6) categories
	catIDs, err := repo.FindCategoryIDs(colCats, post.ID, ctx)
	if err != nil {
		return out, fmt.Errorf("fetch categories: %w", err)
	}
	// 7) is_like
	isLiked, err := repo.CheckIsLiked(ctx, colLikes, loginUserID, post.ID, "post")

	// 8) map -> response
	out = dto.PostResponse{
		UserID:       post.UserID.Hex(),
		Name:         fullName,
		Username:     user.Username,
		PostText:     post.PostText,
		Media:        post.Media,
		Hashtag:      post.Hashtag,
		LikeCount:    post.LikeCount,
		CommentCount: post.CommentCount,
		PostAs: dto.PostAs{
			OrgPath:     orgPath,
			PositionKey: posName,
			Tag:         post.Tags,
		},
		CategoryIDs:  catIDs,
		Visibility:   vis,
		OrgOfContent: orgPath,
		CreatedAt:    post.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:    post.UpdatedAt.UTC().Format(time.RFC3339),
		Status:       post.Status,
		Isliked:   isLiked,
	}
	return out, nil
}
