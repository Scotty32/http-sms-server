package requests

import (
	"strings"

	"github.com/google/uuid"

	"github.com/NdoleStudio/httpsms/pkg/entities"
	"github.com/NdoleStudio/httpsms/pkg/services"
)

// PhoneOperatorsUpdate is the payload for configuring the network operators a phone supports.
// This is an owner/admin-only configuration endpoint, distinct from the Android app's
// self-registration flow (PhoneUpsert) - it lets an integration like OMCI declare which SIM
// serves which network operator so /v2/send can route "strict" requests correctly.
type PhoneOperatorsUpdate struct {
	request
	PhoneID            string   `json:"phoneID" swaggerignore:"true"` // taken from the route param, not the body
	SupportedOperators []string `json:"supported_operators" example:"orange,mtn"`
}

// Sanitize trims and lowercases the operator values
func (input *PhoneOperatorsUpdate) Sanitize() PhoneOperatorsUpdate {
	cleaned := make([]string, 0, len(input.SupportedOperators))
	for _, operator := range input.SupportedOperators {
		operator = strings.ToLower(strings.TrimSpace(operator))
		if operator != "" {
			cleaned = append(cleaned, operator)
		}
	}
	input.SupportedOperators = cleaned
	return *input
}

// PhoneIDUuid returns the phoneID as uuid.UUID
func (input *PhoneOperatorsUpdate) PhoneIDUuid() uuid.UUID {
	return uuid.MustParse(input.PhoneID)
}

// ToUpdateOperatorsParams converts PhoneOperatorsUpdate to services.PhoneOperatorsUpdateParams
func (input *PhoneOperatorsUpdate) ToUpdateOperatorsParams(userID entities.UserID, phoneID uuid.UUID, source string) *services.PhoneOperatorsUpdateParams {
	return &services.PhoneOperatorsUpdateParams{
		Source:             source,
		UserID:             userID,
		PhoneID:            phoneID,
		SupportedOperators: input.SupportedOperators,
	}
}
