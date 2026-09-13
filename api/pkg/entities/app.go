package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// App represents an application that can send SMS messages using one of its linked phone numbers
type App struct {
	ID            uuid.UUID      `json:"id" gorm:"primaryKey;type:uuid;" example:"32343a19-da5e-4b1b-a767-3298a73703cb"`
	UserID        UserID         `json:"user_id" example:"WB7DRDWrJZRGbYrv2CKGkqbzvqdC"`
	UserEmail     string         `json:"user_email" example:"user@gmail.com"`
	Name          string         `json:"name" example:"My Application"`
	APIKey        string         `json:"api_key" gorm:"uniqueIndex:idx_apps_api_key;NOT NULL" example:"ak_DGW8NwQp7mxKaSZ72Xq9v6xxxxx"`
	APISecret     string         `json:"-" gorm:"NOT NULL"`
	PhoneNumbers  pq.StringArray `json:"phone_numbers" example:"+18005550199,+18005550100" gorm:"type:text[]" swaggertype:"array,string"`
	WebhookURL    *string        `json:"webhook_url" example:"https://example.com/webhook" validate:"optional"`
	WebhookSecret *string        `json:"-" gorm:"column:webhook_secret"`
	CreatedAt     time.Time      `json:"created_at" example:"2022-06-05T14:26:02.302718+03:00"`
	UpdatedAt     time.Time      `json:"updated_at" example:"2022-06-05T14:26:02.302718+03:00"`
}

// TableName overrides the table name used by App
func (App) TableName() string {
	return "apps"
}
