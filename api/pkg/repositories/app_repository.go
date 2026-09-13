package repositories

import (
	"context"

	"github.com/NdoleStudio/httpsms/pkg/entities"
	"github.com/google/uuid"
)

// AppRepository loads and persists an entities.App
type AppRepository interface {
	// Create a new entities.App
	Create(ctx context.Context, app *entities.App) error

	// Load an entities.App by userID and appID
	Load(ctx context.Context, userID entities.UserID, appID uuid.UUID) (*entities.App, error)

	// LoadByAPIKey loads an entities.App by its API key
	LoadByAPIKey(ctx context.Context, apiKey string) (*entities.App, error)

	// Index returns all entities.App for a user
	Index(ctx context.Context, userID entities.UserID, params IndexParams) ([]*entities.App, error)

	// Delete an entities.App
	Delete(ctx context.Context, app *entities.App) error

	// Update an entities.App
	Update(ctx context.Context, app *entities.App) error

	// AddPhoneNumber adds a phone number to an entities.App
	AddPhoneNumber(ctx context.Context, app *entities.App, phoneNumber string) error

	// RemovePhoneNumber removes a phone number from an entities.App
	RemovePhoneNumber(ctx context.Context, app *entities.App, phoneNumber string) error

	// DeleteAllForUser deletes all entities.App for a user
	DeleteAllForUser(ctx context.Context, userID entities.UserID) error
}
