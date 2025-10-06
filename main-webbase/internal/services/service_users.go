package services

import (
	"context"

	"main-webbase/dto"
	repo "main-webbase/internal/repository"
)

func GetUserProfile(ctx context.Context, userID string) (*dto.UserProfileDTO, error) {
	users, err := repo.FindUserBy(ctx, "_id", userID)
	if err != nil {
		return nil, err
	}

	user := users[0]
	memberships, err := repo.GetUserMemberships(ctx, user.ID.Hex())
	if err != nil {
		return nil, err
	}

	var membershipDetails []dto.MembershipProfileDTO
	for _, m := range memberships {
		org, err := repo.FindByOrgPath(ctx, m.OrgPath)
		if err != nil {
			return nil, err
		}
		pos, err := repo.FindPositionByKeyandPath(ctx, m.PositionKey, m.OrgPath)
		if err != nil {
			return nil, err
		}
		policies, err := repo.FindPolicyByKeyandPath(ctx, m.PositionKey, m.OrgPath)
		if err != nil {
			return nil, err
		}

		membershipDetails = append(membershipDetails, dto.MembershipProfileDTO{
			MembershipName: pos.Display["en"], // or pos.Display["th"] depending on locale
			OrgUnit:        *org,
			Position:       *pos,
			Policies:       *policies,
		})
	}

	userprofile := &dto.UserProfileDTO{
		ID:          user.ID.Hex(),
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Email:       user.Email,
		ThaiPrefix:  user.ThaiPrefix,
		Gender:      user.Gender,
		TypePerson:  user.TypePerson,
		StudentID:   user.StudentID,
		AdvisorID:   user.AdvisorID,
		Memberships: membershipDetails,
	}

	return userprofile, nil
}