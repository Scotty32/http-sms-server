package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/NdoleStudio/httpsms/pkg/entities"
	"github.com/NdoleStudio/httpsms/pkg/telemetry"
	"github.com/palantir/stacktrace"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// gormPlanRepository is responsible for persisting entities.Plan
type gormPlanRepository struct {
	logger telemetry.Logger
	tracer telemetry.Tracer
	db     *gorm.DB
}

// NewGormPlanRepository creates the GORM version of the PlanRepository
func NewGormPlanRepository(
	logger telemetry.Logger,
	tracer telemetry.Tracer,
	db *gorm.DB,
) PlanRepository {
	return &gormPlanRepository{
		logger: logger.WithService(fmt.Sprintf("%T", &gormPlanRepository{})),
		tracer: tracer,
		db:     db,
	}
}

// Upsert creates or updates an entities.Plan by its unique Name
func (repository *gormPlanRepository) Upsert(ctx context.Context, plan *entities.Plan) error {
	ctx, span := repository.tracer.Start(ctx)
	defer span.End()

	err := repository.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoUpdates: clause.AssignmentColumns([]string{"display_name", "message_limit", "phone_limit", "price_cents", "currency", "billing_period", "updated_at"}),
	}).Create(plan).Error
	if err != nil {
		msg := fmt.Sprintf("cannot upsert [%T] with name [%s]", plan, plan.Name)
		return repository.tracer.WrapErrorSpan(span, stacktrace.Propagate(err, msg))
	}

	return nil
}

// LoadByName fetches an entities.Plan by its unique Name
func (repository *gormPlanRepository) LoadByName(ctx context.Context, name string) (*entities.Plan, error) {
	ctx, span := repository.tracer.Start(ctx)
	defer span.End()

	plan := new(entities.Plan)
	err := repository.db.WithContext(ctx).Where("name = ?", name).First(plan).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		msg := fmt.Sprintf("plan with name [%s] does not exist", name)
		return nil, repository.tracer.WrapErrorSpan(span, stacktrace.PropagateWithCode(err, ErrCodeNotFound, msg))
	}
	if err != nil {
		msg := fmt.Sprintf("cannot load [%T] with name [%s]", plan, name)
		return nil, repository.tracer.WrapErrorSpan(span, stacktrace.Propagate(err, msg))
	}

	return plan, nil
}

// Index fetches all entities.Plan
func (repository *gormPlanRepository) Index(ctx context.Context) ([]*entities.Plan, error) {
	ctx, span := repository.tracer.Start(ctx)
	defer span.End()

	var plans []*entities.Plan
	if err := repository.db.WithContext(ctx).Order("price_cents ASC").Find(&plans).Error; err != nil {
		return nil, repository.tracer.WrapErrorSpan(span, stacktrace.Propagate(err, fmt.Sprintf("cannot index [%T]", entities.Plan{})))
	}

	return plans, nil
}
