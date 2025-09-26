package services

import (
	"context"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"main-webbase/database"
	"main-webbase/internal/models"
)

func AllUserOrg(userID bson.ObjectID) ([]string, error) {
	collection_membership := database.DB.Collection("memberships")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection_membership.Find(ctx, bson.M{"user_id": userID, "active": true})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	organize_set := map[string]struct{}{}

	for cursor.Next(ctx) {
		var user_org models.MembershipDoc
		if err := cursor.Decode(&user_org); err != nil {
			return nil, err
		}
		if user_org.OrgPath == "" {
			continue
		}
		organize_set[user_org.OrgPath] = struct{}{}

		parts := strings.Split(user_org.OrgPath, "/")
		for i := 1; i < len(parts); i++ {
			parent := strings.Join(parts[:i], "/")
			if parent != "" {
				organize_set[parent] = struct{}{}
			}
		}
	}

	// Flatten
	orgs := make([]string, 0, len(organize_set))
	for path := range organize_set {
		orgs = append(orgs, path)
	}
	return orgs, nil
}
