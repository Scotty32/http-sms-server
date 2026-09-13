package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/NdoleStudio/httpsms/pkg/entities"
	"github.com/NdoleStudio/httpsms/pkg/telemetry"
	"github.com/dgraph-io/ristretto/v2"
	"github.com/google/uuid"
	"github.com/palantir/stacktrace"
	"gorm.io/gorm"
)

// gormAppRepository is responsible for persisting entities.App
type gormAppRepository struct {
	logger telemetry.Logger
	tracer telemetry.Tracer
	cache  *ristretto.Cache[string, *entities.App]
	db     *gorm.DB
}

// NewGormAppRepository creates the GORM version of the AppRepository
func NewGormAppRepository(
	logger telemetry.Logger,
	tracer telemetry.Tracer,
	db *gorm.DB,
	cache *ristretto.Cache[string, *entities.App],
) AppRepository {
	return &gormAppRepository{
		logger: logger.WithService(fmt.Sprintf("%T", &gormAppRepository{})),
		tracer: tracer,
		cache:  cache,
		db:     db,
	}
}

// Create a new entities.App
func (repository *gormAppRepository) Create(ctx context.Context, app *entities.App) error {
	ctx, span := repository.tracer.Start(ctx)
	defer span.End()

	if err := repository.db.WithContext(ctx).Create(app).Error; err != nil {
		msg := fmt.Sprintf("cannot create app with ID [%s] for user with ID [%s]", app.ID, app.UserID)
		return repository.tracer.WrapErrorSpan(span, stacktrace.Propagate(err, msg))
	}

	return nil
}

// Load an entities.App by userID and appID
func (repository *gormAppRepository) Load(ctx context.Context, userID entities.UserID, appID uuid.UUID) (*entities.App, error) {
	ctx, span := repository.tracer.Start(ctx)
	defer span.End()

	app := new(entities.App)
	err := repository.db.WithContext(ctx).Where("user_id = ?", userID).Where("id = ?", appID).First(app).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		msg := fmt.Sprintf("[%T] with ID [%s] for user with ID [%s] does not exist", app, appID, userID)
		return nil, repository.tracer.WrapErrorSpan(span, stacktrace.PropagateWithCode(err, ErrCodeNotFound, msg))
	}

	if err != nil {
		msg := fmt.Sprintf("cannot load [%T] with ID [%s] for user with ID [%s]", app, appID, userID)
		return nil, repository.tracer.WrapErrorSpan(span, stacktrace.Propagate(err, msg))
	}

	return app, nil
}

// LoadByAPIKey loads an entities.App by its API key
func (repository *gormAppRepository) LoadByAPIKey(ctx context.Context, apiKey string) (*entities.App, error) {
	ctx, span, ctxLogger := repository.tracer.StartWithLogger(ctx, repository.logger)
	defer span.End()

	if cached, found := repository.cache.Get(apiKey); found {
		return cached, nil
	}

	app := new(entities.App)
	err := repository.db.WithContext(ctx).Where("api_key = ?", apiKey).First(app).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		msg := fmt.Sprintf("app with api key [%s] does not exist", apiKey)
		return nil, repository.tracer.WrapErrorSpan(span, stacktrace.PropagateWithCode(err, ErrCodeNotFound, msg))
	}

	if err != nil {
		msg := fmt.Sprintf("cannot load app with api key [%s]", apiKey)
		return nil, repository.tracer.WrapErrorSpan(span, stacktrace.Propagate(err, msg))
	}

	if result := repository.cache.SetWithTTL(apiKey, app, 1, 30*time.Second); !result {
		ctxLogger.Error(repository.tracer.WrapErrorSpan(span, stacktrace.NewError(fmt.Sprintf("cannot cache app with api key [%s]", apiKey))))
	}

	return app, nil
}

// Index returns all entities.App for a user
func (repository *gormAppRepository) Index(ctx context.Context, userID entities.UserID, params IndexParams) ([]*entities.App, error) {
	ctx, span := repository.tracer.Start(ctx)
	defer span.End()

	query := repository.db.WithContext(ctx).Where("user_id = ?", userID)
	if len(params.Query) > 0 {
		queryPattern := "%" + params.Query + "%"
		query = query.Where("name ILIKE ?", queryPattern)
	}

	apps := new([]*entities.App)
	if err := query.Order("created_at DESC").Limit(params.Limit).Offset(params.Skip).Find(apps).Error; err != nil {
		msg := fmt.Sprintf("cannot fetch apps with userID [%s] and params [%+#v]", userID, params)
		return nil, repository.tracer.WrapErrorSpan(span, stacktrace.Propagate(err, msg))
	}

	return *apps, nil
}

// Delete an entities.App
func (repository *gormAppRepository) Delete(ctx context.Context, app *entities.App) error {
	ctx, span := repository.tracer.Start(ctx)
	defer span.End()

	if err := repository.db.WithContext(ctx).Delete(app).Error; err != nil {
		msg := fmt.Sprintf("cannot delete app with ID [%s] and userID [%s]", app.ID, app.UserID)
		return repository.tracer.WrapErrorSpan(span, stacktrace.Propagate(err, msg))
	}

	repository.cache.Del(app.APIKey)
	return nil
}

// Update an entities.App
func (repository *gormAppRepository) Update(ctx context.Context, app *entities.App) error {
	ctx, span := repository.tracer.Start(ctx)
	defer span.End()

	if err := repository.db.WithContext(ctx).Save(app).Error; err != nil {
		msg := fmt.Sprintf("cannot update app with ID [%s] for user with ID [%s]", app.ID, app.UserID)
		return repository.tracer.WrapErrorSpan(span, stacktrace.Propagate(err, msg))
	}

	repository.cache.Del(app.APIKey)
	return nil
}

// AddPhoneNumber adds a phone number to an entities.App
func (repository *gormAppRepository) AddPhoneNumber(ctx context.Context, app *entities.App, phoneNumber string) error {
	ctx, span := repository.tracer.Start(ctx)
	defer span.End()

	query := `
UPDATE apps
SET phone_numbers = array_append(phone_numbers, ?)
WHERE id = ? AND array_position(phone_numbers, ?) IS NULL;
`
	if err := repository.db.WithContext(ctx).Exec(query, phoneNumber, app.ID, phoneNumber).Error; err != nil {
		msg := fmt.Sprintf("cannot add phone number [%s] to app with ID [%s]", phoneNumber, app.ID)
		return repository.tracer.WrapErrorSpan(span, stacktrace.Propagate(err, msg))
	}

	repository.cache.Del(app.APIKey)
	return nil
}

// RemovePhoneNumber removes a phone number from an entities.App
func (repository *gormAppRepository) RemovePhoneNumber(ctx context.Context, app *entities.App, phoneNumber string) error {
	ctx, span := repository.tracer.Start(ctx)
	defer span.End()

	query := `
UPDATE apps
SET phone_numbers = array_remove(phone_numbers, ?)
WHERE id = ?;
`
	if err := repository.db.WithContext(ctx).Exec(query, phoneNumber, app.ID).Error; err != nil {
		msg := fmt.Sprintf("cannot remove phone number [%s] from app with ID [%s]", phoneNumber, app.ID)
		return repository.tracer.WrapErrorSpan(span, stacktrace.Propagate(err, msg))
	}

	repository.cache.Del(app.APIKey)
	return nil
}

// DeleteAllForUser deletes all entities.App for a user
func (repository *gormAppRepository) DeleteAllForUser(ctx context.Context, userID entities.UserID) error {
	ctx, span := repository.tracer.Start(ctx)
	defer span.End()

	if err := repository.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&entities.App{}).Error; err != nil {
		msg := fmt.Sprintf("cannot delete all [%T] for user with ID [%s]", &entities.App{}, userID)
		return repository.tracer.WrapErrorSpan(span, stacktrace.Propagate(err, msg))
	}

	repository.cache.Clear()
	return nil
}
