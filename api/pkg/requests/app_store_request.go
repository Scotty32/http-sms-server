package requests

import "strings"

// AppStoreRequest is the payload for creating a new App
type AppStoreRequest struct {
	request
	Name          string  `json:"name" example:"My Application"`
	WebhookURL    *string `json:"webhook_url" example:"https://example.com/webhook" validate:"optional"`
	WebhookSecret *string `json:"webhook_secret" example:"my-webhook-secret" validate:"optional"`
}

// Sanitize sets defaults to AppStoreRequest
func (input *AppStoreRequest) Sanitize() AppStoreRequest {
	input.Name = strings.TrimSpace(input.Name)
	if input.WebhookURL != nil {
		sanitized := input.sanitizeURL(*input.WebhookURL)
		input.WebhookURL = &sanitized
	}
	if input.WebhookSecret != nil {
		sanitized := strings.TrimSpace(*input.WebhookSecret)
		input.WebhookSecret = &sanitized
	}
	return *input
}
