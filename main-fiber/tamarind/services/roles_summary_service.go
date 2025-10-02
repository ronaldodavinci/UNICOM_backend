package services

import (
	"context"
	"fmt"
	"time"

	"github.com/pllus/main-fiber/tamarind/repositories"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RolesSummaryService struct {
	membershipRepo *repositories.MembershipRepository
	positionRepo   *repositories.PolicyRepository
	orgRepo        *repositories.OrgUnitRepository
	userRepo       *repositories.UserRepository
}

func NewRolesSummaryService(m *repositories.MembershipRepository, p *repositories.PolicyRepository, o *repositories.OrgUnitRepository, u *repositories.UserRepository) *RolesSummaryService {
	return &RolesSummaryService{
		membershipRepo: m,
		positionRepo:   p,
		orgRepo:        o,
		userRepo:       u,
	}
}

func (s *RolesSummaryService) UpdateRolesSummary(ctx context.Context, userID primitive.ObjectID) error {
	mems, err := s.membershipRepo.FindByUser(ctx, userID)
	if err != nil {
		return err
	}

	type Item struct{ OrgPath, PositionKey, Label string }
	var items []Item
	var orgs []string
	var pos []string

	for _, m := range mems {
		orgs = append(orgs, m.OrgPath)
		pos = append(pos, m.PositionKey)
		items = append(items, Item{
			OrgPath:     m.OrgPath,
			PositionKey: m.PositionKey,
			Label:       fmt.Sprintf("%s • %s", m.PositionKey, m.OrgPath),
		})
	}

	now := time.Now()
	update := bson.M{
		"roles_summary": bson.M{
			"updated_at":    now,
			"memberships":   items,
			"org_paths":     orgs,
			"position_keys": pos,
		},
		"updatedAt": now,
	}
	return s.userRepo.UpdateRolesSummary(ctx, userID, update)
}
