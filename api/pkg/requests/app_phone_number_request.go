package requests

// AppPhoneNumberRequest is the payload for adding/removing a phone number from an App
type AppPhoneNumberRequest struct {
	request
	PhoneNumber string `json:"phone_number" example:"+18005550199"`
}

// Sanitize sets defaults to AppPhoneNumberRequest
func (input *AppPhoneNumberRequest) Sanitize() AppPhoneNumberRequest {
	input.PhoneNumber = input.sanitizeAddress(input.PhoneNumber)
	return *input
}
