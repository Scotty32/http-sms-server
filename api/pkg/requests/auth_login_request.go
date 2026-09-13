package requests

import "strings"

// AuthLoginRequest is the payload for user login
type AuthLoginRequest struct {
	request
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"my-secret-password"`
}

// Sanitize cleans the request inputs
func (input *AuthLoginRequest) Sanitize() AuthLoginRequest {
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	return *input
}
