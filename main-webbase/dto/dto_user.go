package dto

import (
	"main-webbase/internal/models"
)

type MembershipProfileDTO struct {
	MembershipName string           `json:"membership_name"` // e.g., Position Display Name
	OrgUnit        models.OrgUnitNode `json:"org_unit"`
	Position       models.Position    `json:"position"`
	Policies       models.Policy    `json:"policies"`
}

type UserProfileDTO struct {
	ID         string                   `json:"id"`
	FirstName  string                   `json:"firstname"`
	LastName   string                   `json:"lastname"`
	Email      string                   `json:"email"`
	ThaiPrefix string                   `json:"thaiprefix,omitempty"`
	Gender     string                   `json:"gender,omitempty"`
	TypePerson string                   `json:"type_person,omitempty"`
	StudentID  string                   `json:"student_id,omitempty"`
	AdvisorID  string                   `json:"advisor_id,omitempty"`
	Memberships []MembershipProfileDTO  `json:"memberships"`
}