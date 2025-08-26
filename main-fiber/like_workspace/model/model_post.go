package model

type Post struct {
    ID        string   `bson:"_id,omitempty" json:"_id"`
    LikedBy   []string `bson:"likedBy" json:"likedBy"`
    Likes     int      `bson:"likes" json:"likes"`
    Message   string   `bson:"message" json:"message"`
    Name      string   `bson:"name" json:"name"`
    Timestamp string   `bson:"timestamp" json:"timestamp"`
    UID       string   `bson:"uid" json:"uid"`
    Username  string   `bson:"username" json:"username"`
}
