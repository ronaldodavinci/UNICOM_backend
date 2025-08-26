package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

type User struct {
	ID          bson.ObjectID      `bson:"_id,omitempty" json:"id"`
	FirstName   string             `bson:"user_firstname" json:"first_name"`
	LastName    string             `bson:"user_lastname"  json:"last_name"`
	ThaiPrename string             `bson:"thaiprename"    json:"thaiprename"`
	Gender      string             `bson:"gender"         json:"gender"`
	TypePerson  string             `bson:"type_person"    json:"type_person"`
	StudentID   string             `bson:"student_id"     json:"student_id"`
	AdvisorID   string             `bson:"advisor_id"     json:"advisor_id"`
}
