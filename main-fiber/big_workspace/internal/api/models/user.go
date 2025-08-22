package models

import (
	"go.mongodb.org/mongo-driver/v2/bson/primitive"
)

type User struct {
	ID 	  primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name  string             `bson:"name" json:"name"`
	Email string             `bson:"email" json:"email"`
}