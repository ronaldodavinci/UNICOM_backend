package dto

type PolicyUpdateDTO struct {
	OrgPath 	string   `json:"org_path"`
	Key    		string 	 `json:"key"`
	Actions   	[]string `json:"actions"`
	Enabled   	bool     `json:"enabled"`
}