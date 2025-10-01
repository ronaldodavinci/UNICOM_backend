package services

import (
	"context"

	"github.com/pllus/main-fiber/tamarind/repositories"
)

type Audience struct {
	ExactOrgPaths []string
	AudienceKeys  []string
}

type AudienceService struct {
	membershipRepo *repositories.MembershipRepository
}

func NewAudienceService(m *repositories.MembershipRepository) *AudienceService {
	return &AudienceService{membershipRepo: m}
}

func (s *AudienceService) GetAudience(ctx context.Context, userID any) (Audience, error) {
	mems, err := s.membershipRepo.FindByUser(ctx, userID)
	if err != nil {
		return Audience{}, err
	}

	setExact := map[string]struct{}{}
	setKeys := map[string]struct{}{}

	for _, m := range mems {
		setExact[m.OrgPath] = struct{}{}
		setKeys[m.OrgPath] = struct{}{}
		// TODO: if org_ancestors are needed, preload in repo or compute here
	}

	var exact, keys []string
	for k := range setExact {
		exact = append(exact, k)
	}
	for k := range setKeys {
		keys = append(keys, k)
	}
	return Audience{ExactOrgPaths: exact, AudienceKeys: keys}, nil
}
