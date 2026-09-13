package requests

import "strings"

// AuthRegisterRequest is the payload for user registration
type AuthRegisterRequest struct {
	request
	Name     string `json:"name" example:"John Doe"`
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"my-secret-password"`
}

// Sanitize cleans the request inputs
func (input *AuthRegisterRequest) Sanitize() AuthRegisterRequest {
	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	return *input
}
