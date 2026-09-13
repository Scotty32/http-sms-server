package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/NdoleStudio/httpsms/pkg/entities"
	"github.com/NdoleStudio/httpsms/pkg/telemetry"
	"github.com/google/uuid"
	"github.com/palantir/stacktrace"
	"gorm.io/gorm"
)

// gormSessionRepository is responsible for persisting entities.Session
type gormSessionRepository struct {
	logger telemetry.Logger
	tracer telemetry.Tracer
	db     *gorm.DB
}

// NewGormSessionRepository creates the GORM version of the SessionRepository
func NewGormSessionRepository(
	logger telemetry.Logger,
	tracer telemetry.Tracer,
	db *gorm.DB,
) SessionRepository {
	return &gormSessionRepository{
		logger: logger.WithService(fmt.Sprintf("%T", &gormSessionRepository{})),
		tracer: tracer,
		db:     db,
	}
}

// Store a new entities.Session
func (r *gormSessionRepository) Store(ctx context.Context, session *entities.Session) error {
	ctx, span := r.tracer.Start(ctx)
	defer span.End()

	if err := r.db.WithContext(ctx).Create(session).Error; err != nil {
		return r.tracer.WrapErrorSpan(span, stacktrace.Propagate(err, fmt.Sprintf("cannot store session with ID [%s]", session.ID)))
	}
	return nil
}

// FindByID loads an entities.Session by its ID
func (r *gormSessionRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Session, error) {
	ctx, span := r.tracer.Start(ctx)
	defer span.End()

	session := new(entities.Session)
	err := r.db.WithContext(ctx).Where("id = ?", id).First(session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, r.tracer.WrapErrorSpan(span, stacktrace.PropagateWithCode(err, ErrCodeNotFound, fmt.Sprintf("session with ID [%s] does not exist", id)))
	}
	if err != nil {
		return nil, r.tracer.WrapErrorSpan(span, stacktrace.Propagate(err, fmt.Sprintf("cannot find session with ID [%s]", id)))
	}
	return session, nil
}

// DeleteByID deletes an entities.Session by its ID
func (r *gormSessionRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	ctx, span := r.tracer.Start(ctx)
	defer span.End()

	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&entities.Session{}).Error; err != nil {
		return r.tracer.WrapErrorSpan(span, stacktrace.Propagate(err, fmt.Sprintf("cannot delete session with ID [%s]", id)))
	}
	return nil
}

// DeleteExpired deletes all expired sessions
func (r *gormSessionRepository) DeleteExpired(ctx context.Context) error {
	ctx, span := r.tracer.Start(ctx)
	defer span.End()

	if err := r.db.WithContext(ctx).Where("expires_at < ?", time.Now().UTC()).Delete(&entities.Session{}).Error; err != nil {
		return r.tracer.WrapErrorSpan(span, stacktrace.Propagate(err, "cannot delete expired sessions"))
	}
	return nil
}
