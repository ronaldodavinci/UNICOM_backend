package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"time"
)

type User struct {
	ID         bson.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	FirstName  string        `bson:"firstname" json:"firstname"`
	LastName   string        `bson:"lastname" json:"lastname"`
	ThaiPrefix string        `bson:"thaiprefix,omitempty" json:"thaiprefix,omitempty"`
	Gender     string        `bson:"gender,omitempty" json:"gender,omitempty"`
	TypePerson string        `bson:"type_person,omitempty" json:"type_person,omitempty"`
	StudentID  string        `bson:"student_id,omitempty" json:"student_id,omitempty"`
	AdvisorID  string        `bson:"advisor_id,omitempty" json:"advisor_id,omitempty"`
	// profile_pic string        `bson:"profile_pic,omitempty" json:"profile_pic,omitempty"`

	// ADD เพิ่ม
	Email        string    `bson:"email" json:"email"`
	PasswordHash string    `bson:"password_hash,omitempty" json:"-"`
	CreatedAt    time.Time `bson:"createdAt,omitempty" json:"createdAt,omitempty"`
	UpdatedAt    time.Time `bson:"updatedAt,omitempty" json:"updatedAt,omitempty"`
}

// User
// UserFirstname
// UserLastname
// Othercomponent...
// Membership 1 (cspk member) {
// Orgunit detail
// Position detail
// Policy detail
// }
// Membership 2 (comengineer student) {
// Orgunit detail
// Position detail
// Policy detail
// }
