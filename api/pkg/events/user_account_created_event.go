package events

import (
	"time"

	"github.com/NdoleStudio/httpsms/pkg/entities"
)

// UserAccountCreated is raised when a user's account is created.
const UserAccountCreated = "user.account.created"

// UserAccountCreatedPayload stores the data for the UserAccountCreated event
type UserAccountCreatedPayload struct {
	UserID    entities.UserID `json:"user_id"`
	UserEmail string          `json:"user_email"`
	Timestamp time.Time       `json:"timestamp"`
}
