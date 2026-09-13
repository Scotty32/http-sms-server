package entities

import "time"

// Plan represents a subscription plan/pricing tier that can be assigned to a User.
// Plans are seeded/updated idempotently from a JSON catalog via the plans CLI command
// (cmd/plans) rather than edited directly - Name is the stable key referenced by
// User.SubscriptionName and by LemonSqueezy variant mapping.
type Plan struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	Name          string    `json:"name" gorm:"uniqueIndex:idx_plans_name;NOT NULL" example:"free"`
	DisplayName   string    `json:"display_name" example:"Free"`
	MessageLimit  int       `json:"message_limit" example:"200"`
	PhoneLimit    int       `json:"phone_limit" example:"1"`
	PriceCents    int       `json:"price_cents" example:"999"`
	Currency      string    `json:"currency" example:"USD"`
	BillingPeriod string    `json:"billing_period" example:"monthly"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// TableName overrides the table name used by Plan
func (Plan) TableName() string {
	return "plans"
}
