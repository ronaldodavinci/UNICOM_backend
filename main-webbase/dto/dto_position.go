package dto

import (
	"main-webbase/internal/models"
)

type PositionCreateDTO struct {
	Key         string             `bson:"key" json:"key"`
	Display     map[string]string  `bson:"display" json:"display"`
	Rank        int                `bson:"rank" json:"rank"`
	Scope       *models.Scope           `bson:"scope" json:"scope"`
	Constraints *models.Constraints     `bson:"constraints" json:"constraints"`
	Status      string             `bson:"status" json:"status"`

	Policy struct {
		Type        string    `bson:"type" json:"type"` // "exact" or "subtree"
		Actions     []string  `bson:"actions" json:"actions"`
		Enabled     bool      `bson:"enabled" json:"enabled"`
	} `bson:"policy" json:"policy"`
}