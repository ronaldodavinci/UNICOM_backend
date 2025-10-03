package dto

// Response node for org tree
type OrgUnitNode struct {
	OrgPath   string            `json:"org_path"`
	Type      string            `json:"type,omitempty"`
	Label     string            `json:"label,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
	ShortName string            `json:"short_name,omitempty"`
	Children  []*OrgUnitNode    `json:"children,omitempty"`
}
