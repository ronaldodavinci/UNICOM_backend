package controllers

import (
	"context"
	"errors"
	"fmt"
	"main-webbase/dto"
	"main-webbase/internal/accessctx"
	mid "main-webbase/internal/middleware"
	repo "main-webbase/internal/repository"
	"main-webbase/internal/services"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func canPostAs(v *accessctx.ViewerAccess, orgPath, positionKey string) bool {
	if v == nil {
		return false
	}
	// root/admin (OrgPath == "/") ใช้ได้ทุกอย่าง
	if isRootByPath(v) {
		return true
	}
	// ต้องมี membership ที่ตรงทั้ง org_path และ position_key
	for _, m := range v.Memberships {
		// หมายเหตุ: field ชื่อ PosKey ใน viewer
		if m.OrgPath == orgPath && m.PosKey == positionKey {
			return true
		}
	}
	return false
}

// POST /posts

// CreatePostHandler godoc
// @Summary      Create a post
// @Description  Create a new post with categories, visibility and media URLs
// @Tags         posts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header    string             true  "Bearer {token}"
// @Param        data           body      dto.CreatePostDTO  true  "Post payload"
// @Success      201            {object}  dto.PostResponse
// @Failure      400            {object}  dto.ErrorResponse
// @Failure      401            {object}  dto.ErrorResponse
// @Failure      404            {object}  dto.ErrorResponse
// @Failure      500            {object}  dto.ErrorResponse
// @Router       /posts [post]
func CreatePostHandler(client *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, _ := mid.UIDFromLocals(c)
		// fmt.Println("UIDFromLocals: uid=%s\n", uid)

		var body dto.CreatePostDTO
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).
				JSON(dto.ErrorResponse{Error: "invalid body"})
		}

		// --- basic validation ---
		if body.PostText == "" {
			return c.Status(fiber.StatusBadRequest).
				JSON(dto.ErrorResponse{Error: "postText is required"})
		}
		if body.PostAs.OrgPath == "" || body.PostAs.PositionKey == "" {
			return c.Status(fiber.StatusBadRequest).
				JSON(dto.ErrorResponse{Error: "postAs.org_path and postAs.position_key are required"})
		}

		if !canPostAs(viewerFrom(c), body.PostAs.OrgPath, body.PostAs.PositionKey) {
			// fmt.Printf("[DEBUG viewer] %+v\n", m.OrgPath)
			// v := viewerFrom(c)
			// b, _ := json.MarshalIndent(v, "", "  ")
			// fmt.Println("[DEBUG] Viewer =", string(b))
			return c.Status(fiber.StatusForbidden).
				JSON(dto.ErrorResponse{Error: "forbidden: you cannot post as this role1"})
		}

		// default visibility
		if body.Visibility.Access == "" {
			body.Visibility.Access = "public"
		}

		// Use a proper context.Context for Mongo (ไม่ใช้ c.Context() เพราะไม่ใช่ Go context)
		ctx := context.Background()

		post, err := services.CreatePostWithMeta(client, userID, body, ctx)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrUserNotFound):
				return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Error: "user not found1"})
			case errors.Is(err, services.ErrOrgNodeNotFound):
				return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Error: "org_path not found1"})
			case errors.Is(err, services.ErrPositionNotFound):
				return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Error: "position_key not found"})
			default:
				// อาจเช็ค duplicate key, validation อื่น ๆ เพิ่มได้หาก service ตีกลับ error เฉพาะ
				return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Error: err.Error()})
			}
		}

		return c.Status(fiber.StatusCreated).JSON(post)
	}
}

// GET /posts/:post_id

// GetIndividualPostHandler godoc
// @Summary      Get a post detail
// @Description  Return post detail (user, position, org path, visibility, categories, likes count, etc.)
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        post_id  path  string  true  "Post ID (hex)"
// @Success      200  {object}  dto.PostResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /posts/{post_id} [get]
func GetIndividualPostHandler(client *mongo.Client) fiber.Handler {
	const dbName = "unicom"

	return func(c *fiber.Ctx) error {
		postIDHex := c.Params("post_id")
		if postIDHex == "" {
			return fiber.NewError(fiber.StatusBadRequest, "missing post_id in route")
		}
		postID, err := bson.ObjectIDFromHex(postIDHex)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid post_id")
		}

		// context พร้อม timeout
		ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
		defer cancel()

		db := client.Database(dbName)

		resp, err := services.GetPostDetail(ctx, db, postID)
		if err != nil {
			// ถ้าถูก wrap ด้วย %w จาก service จะเช็ค ErrNoDocuments ได้
			if errors.Is(err, mongo.ErrNoDocuments) {
				return fiber.NewError(fiber.StatusNotFound, "post not found")
			}
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		return c.Status(fiber.StatusOK).JSON(resp)
	}
}

// DeletePostHandlerWithClient สร้าง Fiber handler โดยรับ mongo.Client + dbName
// @Summary      Soft delete post (status: active -> inactive)
// @Description  เปลี่ยนสถานะโพสต์จาก active เป็น inactive (soft delete)
// @Tags         posts
// @Param        id   path      string  true  "Post ID (hex)"
// @Success      204
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /posts/{post_id} [delete]
func DeletePostHandler(client *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		uid, _ := mid.UIDObjectID(c)

		idParam := c.Params("post_id")
		postID, err := bson.ObjectIDFromHex(idParam)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid post id")
		}

		isRoot := isRootByPath(viewerFrom(c))

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		db := client.Database("unicom")

		ok, err := repo.DeletePost(db, postID, ctx, uid, isRoot)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		if !ok {
			// ไม่พบ/ไม่ active หรือไม่มีสิทธิ์ (ไม่ใช่เจ้าของและไม่ใช่ root)
			return fiber.NewError(fiber.StatusForbidden, "forbidden")
		}
		return c.SendStatus(fiber.StatusNoContent)
	}
}

// PUT /posts/:id

// UpdatePostHandler godoc
// @Summary      Update a post (full replace of editable fields)
// @Description  Update post text, medias, categories, visibility, and posted-as. Owner can edit content; admin/root can also change status.
// @Tags         posts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header    string                   true  "Bearer {token}"
// @Param        id             path      string                   true  "Post ID (hex)"
// @Param        data           body      dto.UpdatePostFullDTO    true  "Full update payload"
// @Success      200            {object}  models.Post
// @Failure      400            {object}  dto.ErrorResponse
// @Failure      401            {object}  dto.ErrorResponse
// @Failure      403            {object}  dto.ErrorResponse
// @Failure      500            {object}  dto.ErrorResponse
// @Router       /posts/{post_id} [put]
func UpdatePostHandler(client *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		uid, _ := mid.UIDObjectID(c)
		fmt.Printf("[Handler] uid=%s\n", uid.Hex())
		postID, err := bson.ObjectIDFromHex(c.Params("post_id"))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid post id")
		}
		v := viewerFrom(c)
		// b, _ := json.MarshalIndent(v, "", "  ")
		// fmt.Println("[DEBUG] Viewer =", string(b))
		var body dto.UpdatePostFullDTO
		if err := c.BodyParser(&body); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid body")
		}
		if strings.TrimSpace(body.PostText) == "" {
			return fiber.NewError(fiber.StatusBadRequest, "postText required")
		}

		isRoot := isRootByPath(viewerFrom(c)) // root/admin = OrgPath == "/"

		// ✅ เช็คสิทธิ์ postAs ถ้าส่งมาแก้ (เรา require postAs ใน DTO อยู่แล้ว)
		// ถ้าอยากให้ "ไม่บังคับส่ง postAs ทุกครั้ง" ให้เช็คเฉพาะกรณีที่มีค่าใหม่
		if !canPostAs(v, body.PostAs.OrgPath, body.PostAs.PositionKey) {
			return c.Status(fiber.StatusForbidden).
				JSON(dto.ErrorResponse{Error: "forbidden: you cannot post as this role2"})
		}

		ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
		defer cancel()

		db := client.Database("unicom")
		if _, err := services.UpdatePostFull(client, db, postID, uid, isRoot, body, ctx); err != nil {
			msg := err.Error()
			switch {
			case strings.Contains(msg, "forbidden"):
				// fmt.Println("[FORBIDDEN-3] handler caught forbidden:", err)
				return c.Status(403).JSON(dto.ErrorResponse{Error: "forbidden"})
			case strings.Contains(msg, "invalid") || strings.Contains(msg, "org_path not found2") || strings.Contains(msg, "position_key not found"):
				return c.Status(400).JSON(dto.ErrorResponse{Error: msg})
			case strings.Contains(msg, "post not found"):
				return c.Status(404).JSON(dto.ErrorResponse{Error: msg})
			default:
				return c.Status(500).JSON(dto.ErrorResponse{Error: msg})
			}
		}

		resp, err := services.GetPostDetail(ctx, client.Database("unicom"), postID)
		if err != nil {
			return c.Status(500).JSON(dto.ErrorResponse{Error: err.Error()})
		}
		return c.Status(200).JSON(resp)
	}
}
