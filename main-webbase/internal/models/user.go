package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)



// User represents a user in the database.
type User struct {
	ID           bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	FirstName    string             `bson:"first_name,omitempty" json:"first_name,omitempty"`
	LastName     string             `bson:"last_name,omitempty" json:"last_name,omitempty"`
	ThaiPrename  string             `bson:"thaiprename,omitempty" json:"thaiprename,omitempty"`
	Gender       string             `bson:"gender,omitempty" json:"gender,omitempty"`
	TypePerson   string             `bson:"type_person,omitempty" json:"type_person,omitempty"`
	StudentID    string             `bson:"student_id,omitempty" json:"student_id,omitempty"`
	AdvisorID    string             `bson:"advisor_id,omitempty" json:"advisor_id,omitempty"`
	Username     string             `bson:"username,omitempty" json:"username,omitempty"`
	Password     string             `bson:"password,omitempty" json:"password,omitempty"` // 🔑 Add this line
}
// bson คือ ชื่อที่ขึ้นใน Database mongo Ex.
	// {
	//   "firstname": "Alice",
	//   "lastname": "Smith",
	//   "thaiprename": "นางสาว",
	//   "gender": "Female",
	//   "typeperson": "student",
	//   "studentid": "65012345",
	//   "advisorid": "123"
	// }

// json คือ ขื่อที่ใช้ใน API Ex.
	// {   
	//     "first_name": "Alice",
	//     "last_name": "Smith",
	//     "thaiprename": "นางสาว",
	//     "gender": "Female",
	//     "type_person": "student",
	//     "student_id": "65012345",
	//     "advisor_id": "123"
	// }
