package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User document stored in "users"
type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	SeqID        int                `bson:"id" json:"id"` // numeric app id
	FirstName    string             `bson:"firstName" json:"firstName"`
	LastName     string             `bson:"lastName" json:"lastName"`
	ThaiPrefix   string             `bson:"thaiprefix,omitempty" json:"thaiprefix,omitempty"`
	Gender       string             `bson:"gender,omitempty" json:"gender,omitempty"`
	TypePerson   string             `bson:"type_person,omitempty" json:"type_person,omitempty"`
	StudentID    string             `bson:"student_id,omitempty" json:"student_id,omitempty"`
	AdvisorID    string             `bson:"advisor_id,omitempty" json:"advisor_id,omitempty"`
	Email        string             `bson:"email" json:"email"`
	Roles        []string           `bson:"roles" json:"roles"`
	PasswordHash string             `bson:"password_hash,omitempty" json:"-"`
	CreatedAt    time.Time          `bson:"createdAt,omitempty" json:"createdAt,omitempty"`
	UpdatedAt    time.Time          `bson:"updatedAt,omitempty" json:"updatedAt,omitempty"`
}

// UserWithRoleDetails expands user with role details
type UserWithRoleDetails struct {
	User        `bson:",inline"`
	RoleDetails []Role `bson:"roleDetails" json:"roleDetails"`
}
