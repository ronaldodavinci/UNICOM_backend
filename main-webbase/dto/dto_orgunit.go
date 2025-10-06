package dto

type OrgUnitNode struct {
	OrgPath   string            `json:"org_path"`
	Type      string            `json:"type,omitempty"`
	Label     string            `json:"label,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
	ShortName string            `json:"short_name,omitempty"`
	Children  []*OrgUnitNode    `json:"children,omitempty"`
}

type OrgUnitDTO struct {
	ParentPath   string            	`json:"parent_path"`
	Name       	 string 			`json:"name"`
	Slug      	 string             `json:"slug,omitempty"`
	Type       	 string            	`json:"type"`
}

type OrgUnitReport struct {
	OrgID    	string                `json:"org_id"`
	OrgPath 	string                `json:"org_path"`
	Name    	string				  `json:"name"`
	ShortName   string 			  	  `json:"short_name"`      
}

type OrgUnitTreeQuery struct {
	Start string `query:"start"`
	Depth int    `query:"depth"`
}

type OrgUnitTree struct {
	OrgPath   string         `json:"org_path"`
	Type      string         `json:"type,omitempty"`
	Label     string         `json:"label,omitempty"`
	ShortName string         `json:"short_name,omitempty"`
	Children  []*OrgUnitTree `json:"children,omitempty"`
	Sort      int            `json:"-"`
}