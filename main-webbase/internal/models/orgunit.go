package models

// OrgUnit document stored in "org_units"
type OrgUnit struct {
	OrgPath   string            `bson:"org_path"`
	Parent    string            `bson:"parent_path,omitempty"`
	Type      string            `bson:"type,omitempty"`
	Label     map[string]string `bson:"label,omitempty"`
	ShortName string            `bson:"short_name,omitempty"`
	Sort      int               `bson:"sort,omitempty"`
}
