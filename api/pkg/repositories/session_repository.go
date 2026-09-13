package repositories

import (
	"context"

	"github.com/NdoleStudio/httpsms/pkg/entities"
	"github.com/google/uuid"
)

// SessionRepository loads and persists entities.Session
type SessionRepository interface {
	// Store a new entities.Session
	Store(ctx context.Context, session *entities.Session) error

	// FindByID loads an entities.Session by its ID
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Session, error)

	// DeleteByID deletes an entities.Session by its ID
	DeleteByID(ctx context.Context, id uuid.UUID) error

	// DeleteExpired deletes all expired sessions
	DeleteExpired(ctx context.Context) error
}
