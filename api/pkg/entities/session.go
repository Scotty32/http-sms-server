package entities

import (
	"time"

	"github.com/google/uuid"
)

// Session stores an authenticated user session
type Session struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid" json:"id"`
	UserID    UserID    `gorm:"not null;index" json:"user_id"`
	Email     string    `gorm:"not null" json:"email"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// IsExpired returns true if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().UTC().After(s.ExpiresAt)
}
