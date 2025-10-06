package dto

import (
	"main-webbase/internal/models"
	"time"
)

type PositionCreateDTO struct {
	Key         string             `bson:"key" json:"key"`
	Display     map[string]string  `bson:"display" json:"display"`
	Rank        int                `bson:"rank" json:"rank"`
	Scope       *models.Scope           `bson:"scope" json:"scope"`
	Constraints *models.Constraints     `bson:"constraints" json:"constraints"`
	Status      string             `bson:"status" json:"status"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`

	Policy struct {
		Type        string    `bson:"type" json:"type"` // "exact" or "subtree"
		Actions     []string  `bson:"actions" json:"actions"`
		Effect      string    `bson:"effect" json:"effect"`
		Enabled     bool      `bson:"enabled" json:"enabled"`
		CreatedAt   time.Time `bson:"created_at" json:"created_at"`
	} `bson:"policy" json:"policy"`
}