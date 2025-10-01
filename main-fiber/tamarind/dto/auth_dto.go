package dto

// Request payload for login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Response payload after login
type LoginResponse struct {
	AccessToken string `json:"access_token"`
}
