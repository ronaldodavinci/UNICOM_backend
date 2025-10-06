package dto

import "go.mongodb.org/mongo-driver/v2/bson"

// แก้ไข
// UserResponse คือข้อมูลผู้ใช้ที่ส่งกลับให้ frontend
type UserInfoResponse struct {
    ID         bson.ObjectID `bson:"_id" json:"id"`
    Username   string             `bson:"username" json:"username"`
    FirstName  string             `bson:"user_firstname" json:"first_name"`
    LastName   string             `bson:"user_lastname" json:"last_name"`
    Prename    string             `bson:"thaiprename" json:"thai_prename"`
    Email      string             `bson:"email" json:"email"`
    StudentID  *string             `bson:"student_id" json:"student_id"`
    AdvisorID  *string             `bson:"advisor_id" json:"advisor_id"`
    Gender     string             `bson:"gender" json:"gender"`
    TypePerson string             `bson:"type_person" json:"type_person"`
    // ProfilePic string             `bson:"profile_pic,omitempty" json:"profile_pic,omitempty"`
    Status     string             `bson:"status" json:"status"`
}
