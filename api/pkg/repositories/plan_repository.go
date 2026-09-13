package repositories

import (
	"context"

	"github.com/NdoleStudio/httpsms/pkg/entities"
)

// PlanRepository loads and persists an entities.Plan
type PlanRepository interface {
	// Upsert creates or updates an entities.Plan by its unique Name - idempotent, safe to
	// call repeatedly (e.g. from the plans CLI seed command) with the same catalog data.
	Upsert(ctx context.Context, plan *entities.Plan) error

	// LoadByName fetches an entities.Plan by its unique Name
	LoadByName(ctx context.Context, name string) (*entities.Plan, error)

	// Index fetches all entities.Plan
	Index(ctx context.Context) ([]*entities.Plan, error)
}
