package requests

import (
	"strings"
	"time"

	"github.com/NdoleStudio/httpsms/pkg/entities"
	"github.com/NdoleStudio/httpsms/pkg/services"
	"github.com/google/uuid"
	"github.com/nyaruka/phonenumbers"
)

// AppSendMessageRequest is the payload for sending a message via an App
type AppSendMessageRequest struct {
	request
	To        string     `json:"to" example:"+18005550100"`
	Content   string     `json:"content" example:"Hello from my app!"`
	Encrypted bool       `json:"encrypted" example:"false"`
	RequestID string     `json:"request_id" example:"153554b5-ae44-44a0-8f4f-7bbac5657ad4" validate:"optional"`
	SendAt    *time.Time `json:"send_at" example:"2025-12-19T16:39:57-08:00" validate:"optional"`

	// Operator is the network operator (e.g. orange, mtn, moov) the recipient is on. When set,
	// the server prefers a sending phone whose SupportedOperators includes it.
	Operator string `json:"operator" example:"orange" validate:"optional"`

	// Strict forces the server to only use a phone that supports Operator, refusing the send
	// (no message attempted) if none is configured to support it. Requires Operator to be set.
	Strict bool `json:"strict" example:"false"`
}

// Sanitize sets defaults to AppSendMessageRequest
func (input *AppSendMessageRequest) Sanitize() AppSendMessageRequest {
	input.To = input.sanitizeAddress(input.To)
	input.Content = strings.TrimSpace(input.Content)
	input.RequestID = strings.TrimSpace(input.RequestID)
	input.Operator = strings.ToLower(strings.TrimSpace(input.Operator))
	return *input
}

// ToMessageSendParams converts AppSendMessageRequest to services.MessageSendParams using the given from number, userID and appID
func (input *AppSendMessageRequest) ToMessageSendParams(userID entities.UserID, appID *uuid.UUID, from string, source string) services.MessageSendParams {
	owner, _ := phonenumbers.Parse(from, phonenumbers.UNKNOWN_REGION)
	return services.MessageSendParams{
		Source:            source,
		Owner:             owner,
		Encrypted:         input.Encrypted,
		RequestID:         input.sanitizeStringPointer(input.RequestID),
		UserID:            userID,
		AppID:             appID,
		SendAt:            input.SendAt,
		RequestReceivedAt: time.Now().UTC(),
		Contact:           input.sanitizeAddress(input.To),
		Content:           input.Content,
	}
}
