package requests

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/NdoleStudio/httpsms/pkg/services"
)

// UserUpdate is the payload for updating a phone
type UserUpdate struct {
	request
	Timezone      string  `json:"timezone" example:"Europe/Helsinki"`
	ActivePhoneID string  `json:"active_phone_id" example:"32343a19-da5e-4b1b-a767-3298a73703cb"`
	WebhookURL    *string `json:"webhook_url" example:"https://example.com/sms-delivery" validate:"optional"`
}

// Sanitize sets defaults to MessageOutstanding
func (input *UserUpdate) Sanitize() UserUpdate {
	input.ActivePhoneID = strings.TrimSpace(input.ActivePhoneID)
	input.Timezone = strings.TrimSpace(input.Timezone)
	if input.WebhookURL != nil {
		trimmed := strings.TrimSpace(*input.WebhookURL)
		input.WebhookURL = &trimmed
	}
	return *input
}

// ToUpdateParams converts UserUpdate to services.UserUpdateParams
func (input *UserUpdate) ToUpdateParams() services.UserUpdateParams {
	location, err := time.LoadLocation(input.Timezone)
	if err != nil {
		location = time.UTC
	}

	var activePhoneID *uuid.UUID
	if input.ActivePhoneID != "" {
		val := uuid.MustParse(input.ActivePhoneID)
		activePhoneID = &val
	}

	return services.UserUpdateParams{
		ActivePhoneID: activePhoneID,
		Timezone:      location,
		WebhookURL:    input.WebhookURL,
	}
}
