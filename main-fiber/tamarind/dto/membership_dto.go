package dto

// Request to add membership
type CreateMembershipRequest struct {
	UserID      string `json:"user_id"`
	OrgPath     string `json:"org_path"`
	PositionKey string `json:"position_key"`
}
