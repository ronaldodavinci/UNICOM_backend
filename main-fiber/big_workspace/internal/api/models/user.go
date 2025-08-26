package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

type User struct {
	ID          bson.ObjectID      `bson:"_id,omitempty" json:"id"`
	FirstName   string             `bson:"firstname"     json:"first_name"`
	LastName    string             `bson:"lastname"      json:"last_name"`
	ThaiPrename string             `bson:"thaiprename"   json:"thaiprename"`
	Gender      string             `bson:"gender"        json:"gender"`
	TypePerson  string             `bson:"typeperson"    json:"type_person"`
	StudentID   string             `bson:"studentid"     json:"student_id"`
	AdvisorID   string             `bson:"advisorid"     json:"advisor_id"`
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
