package dto
import(
	"time"
)

 type GetPostResponse struct {
    ID              string    `json:"_id"`
    UserID          string    `json:"user_id"`
    Username        string    `json:"username"`
    ProfilePic      string    `json:"profile_pic"`
    Category        string    `json:"category"`
    Message         string    `json:"message"`
    Picture         *string   `json:"picture,omitempty"`
    Video           *string   `json:"video,omitempty"`
    LikeCount       int       `json:"like_count"`
    CommentCount    int       `json:"comment_count"`
    AuthorRoles     []string  `json:"author_roles"`
    VisibilityRoles []string  `json:"visibility_roles"`
    TimeStamp       time.Time `json:"time_stamp"`
    IsLiked         bool      `json:"is_liked"`
}
