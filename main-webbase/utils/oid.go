package utils

import "go.mongodb.org/mongo-driver/v2/bson"

func Oid(hex string) (bson.ObjectID, error) { 
	return bson.ObjectIDFromHex(hex) 
}