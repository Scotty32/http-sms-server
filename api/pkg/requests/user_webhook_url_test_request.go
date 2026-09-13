package requests

// UserWebhookURLTest is the payload for testing a webhook URL before saving it
type UserWebhookURLTest struct {
	request
	WebhookURL string `json:"webhook_url" example:"https://example.com/sms-delivery"`
}

// Sanitize sets defaults to UserWebhookURLTest
func (input *UserWebhookURLTest) Sanitize() UserWebhookURLTest {
	input.WebhookURL = input.sanitizeURL(input.WebhookURL)
	return *input
}
