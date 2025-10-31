package models

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	FirstName  string `bson:"firstname" json:"firstname"`
	LastName   string `bson:"lastname" json:"lastname"`
	ThaiPrefix string `bson:"thaiprefix,omitempty" json:"thaiprefix,omitempty"`
	Gender     string `bson:"gender,omitempty" json:"gender,omitempty"`
	TypePerson string `bson:"type_person,omitempty" json:"type_person,omitempty"`
	StudentID  string `bson:"student_id,omitempty" json:"student_id,omitempty"`
	AdvisorID  string `bson:"advisor_id,omitempty" json:"advisor_id,omitempty"`
	Email      string `bson:"email" json:"email"`
	Password   string `bson:"password" json:"password"`
	OrgPath    string `bson:"organize_path" json:"organize_path"`
	Telephone  string `bson:"telephone" json:"telephone"`
	Disease    string `bson:"disease,omitempty" json:"disease,omitempty"`
	Allergy    string `bson:"allergy,omitempty" json:"allergy,omitempty"`
}
