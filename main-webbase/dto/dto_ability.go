package dto

// Response structure for /abilities
type AbilitiesResponse struct {
	OrgPath   string          `json:"org_path"`
	Abilities map[string]bool `json:"abilities"`
	Version   string          `json:"version,omitempty"`
}
