
package services
import (
	"context"
	"errors"
	"fmt"
	"time"

	"like_workspace/dto"
	"like_workspace/model"
	repo "like_workspace/internal/repository"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func CreatePostWithMeta(client *mongo.Client, body dto.CreatePostDTO, ctx context.Context) (dto.PostResponse, error) {
	db := client.Database("lll_workspace")
	now := time.Now().UTC()

	userID, err := bson.ObjectIDFromHex(body.UserID)
	if err != nil {
		return dto.PostResponse{}, errors.New("invalid userId")
	}
	roleID, err := bson.ObjectIDFromHex(body.RoleID)
	if err != nil {
		return dto.PostResponse{}, errors.New("invalid roleId")
	}

	post := model.Post{
		UserID:    userID,
		RoleID:    roleID,
		PostText:  body.PostText,
		CreatedAt: now,
		UpdatedAt: now,
		LikeCount: 0,
		CommentCount: 0,
	}


	resp := dto.PostResponse{
		UserID:         post.UserID.Hex(),
		RoleID:         post.RoleID.Hex(),
		PostText:       post.PostText,
		CreatedAt:      post.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      post.UpdatedAt.Format(time.RFC3339),
		LikeCount:      post.LikeCount,
		CommentCount:   post.CommentCount,
		//PictureUrl:     body.PictureURL,
		//VideoUrl:       body.VideoURL,
	}

	postsCol := db.Collection("posts")
	res, err := postsCol.InsertOne(ctx, post)
	if err != nil {
		return resp, err
	}
	post.ID = res.InsertedID.(bson.ObjectID)

	if err := repo.InsertHashtags(db, post, body.PostText, ctx); err != nil {
		fmt.Println("⚠️ hashtag insert failed:", err)
	}


	if len(body.CategoryIDs) > 0 {
		if err := repo.InsertCategories(db, post.ID, body.CategoryIDs, ctx); err != nil {
			_, _ = postsCol.DeleteOne(ctx, bson.M{"_id": post.ID})
			return resp, err
		}
	}
	resp.CategoryIDs = body.CategoryIDs

	if len(body.RoleIDs) > 0 {
		if err := repo.InsertRoleVisibility(db, post.ID, body.RoleIDs, ctx); err != nil {
			_, _ = postsCol.DeleteOne(ctx, bson.M{"_id": post.ID})
			_, _ = db.Collection("post_categories").DeleteMany(ctx, bson.M{"post_id": post.ID})
			return resp, err
		}
	}
	resp.VisibleRoleIDs = body.RoleIDs

	return resp, nil
}
