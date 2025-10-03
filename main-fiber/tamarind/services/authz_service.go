package services

import (
	"context"

	"github.com/pllus/main-fiber/tamarind/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuthzService struct {
	membershipRepo *repositories.MembershipRepository
	policyRepo     *repositories.PolicyRepository
}

func NewAuthzService(m *repositories.MembershipRepository, p *repositories.PolicyRepository) *AuthzService {
	return &AuthzService{membershipRepo: m, policyRepo: p}
}

// AbilitiesFor returns allowed actions for a user at orgPath
func (s *AuthzService) AbilitiesFor(ctx context.Context, userID primitive.ObjectID, orgPath string, actions []string) (map[string]bool, error) {
	result := make(map[string]bool)
	for _, act := range actions {
		allowed, err := s.Can(ctx, userID, orgPath, act)
		if err != nil {
			return nil, err
		}
		result[act] = allowed
	}
	return result, nil
}

// Can checks if a user can perform a specific action in an org
func (s *AuthzService) Can(ctx context.Context, userID primitive.ObjectID, orgPath string, action string) (bool, error) {
	// 1. load memberships
	mems, err := s.membershipRepo.FindByUser(ctx, userID)
	if err != nil {
		return false, err
	}

	// 2. collect positions
	posSet := map[string]struct{}{}
	for _, m := range mems {
		posSet[m.PositionKey] = struct{}{}
	}
	var posArr []string
	for k := range posSet {
		posArr = append(posArr, k)
	}

	// 3. load policies
	pols, err := s.policyRepo.FindByPositionsAndAction(ctx, posArr, action)
	if err != nil {
		return false, err
	}

	// 4. match memberships with policies
	for _, m := range mems {
		for _, p := range pols {
			if p.PositionKey == m.PositionKey && (orgPath == "" || orgPath == m.OrgPath) {
				return true, nil
			}
		}
	}
	return false, nil
}
