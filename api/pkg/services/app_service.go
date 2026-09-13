package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/lib/pq"

	"github.com/NdoleStudio/httpsms/pkg/entities"
	"github.com/NdoleStudio/httpsms/pkg/repositories"
	"github.com/NdoleStudio/httpsms/pkg/telemetry"
	"github.com/google/uuid"
	"github.com/palantir/stacktrace"
)

// AppService is responsible for managing entities.App
type AppService struct {
	service
	logger     telemetry.Logger
	tracer     telemetry.Tracer
	repository repositories.AppRepository
}

// NewAppService creates a new AppService
func NewAppService(
	logger telemetry.Logger,
	tracer telemetry.Tracer,
	repository repositories.AppRepository,
) *AppService {
	return &AppService{
		logger:     logger.WithService(fmt.Sprintf("%T", &AppService{})),
		tracer:     tracer,
		repository: repository,
	}
}

// AppStoreParams are the parameters for creating a new App
type AppStoreParams struct {
	Name          string
	WebhookURL    *string
	WebhookSecret *string
}

// Create a new entities.App
func (service *AppService) Create(ctx context.Context, authContext entities.AuthContext, params AppStoreParams) (*entities.App, error) {
	ctx, span, ctxLogger := service.tracer.StartWithLogger(ctx, service.logger)
	defer span.End()

	apiKey, err := service.generateKey(48)
	if err != nil {
		return nil, stacktrace.Propagate(err, "cannot generate API key")
	}

	apiSecret, err := service.generateKey(48)
	if err != nil {
		return nil, stacktrace.Propagate(err, "cannot generate API secret")
	}

	app := &entities.App{
		ID:            uuid.New(),
		UserID:        authContext.ID,
		UserEmail:     authContext.Email,
		Name:          params.Name,
		APIKey:        "ak_" + apiKey,
		APISecret:     "sk_" + apiSecret,
		PhoneNumbers:  pq.StringArray{},
		WebhookURL:    params.WebhookURL,
		WebhookSecret: params.WebhookSecret,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	if err = service.repository.Create(ctx, app); err != nil {
		msg := fmt.Sprintf("cannot create App for user [%s]", authContext.ID)
		return nil, service.tracer.WrapErrorSpan(span, stacktrace.Propagate(err, msg))
	}

	ctxLogger.Info(fmt.Sprintf("created [%T] with ID [%s] for user ID [%s]", app, app.ID, authContext.ID))
	return app, nil
}

// Index fetches all entities.App for a user
func (service *AppService) Index(ctx context.Context, userID entities.UserID, params repositories.IndexParams) ([]*entities.App, error) {
	ctx, span, ctxLogger := service.tracer.StartWithLogger(ctx, service.logger)
	defer span.End()

	apps, err := service.repository.Index(ctx, userID, params)
	if err != nil {
		msg := fmt.Sprintf("could not fetch apps with params [%+#v]", params)
		return nil, service.tracer.WrapErrorSpan(span, stacktrace.Propagate(err, msg))
	}

	ctxLogger.Info(fmt.Sprintf("fetched [%d] apps for user [%s]", len(apps), userID))
	return apps, nil
}

// Load fetches a single entities.App by its ID
func (service *AppService) Load(ctx context.Context, userID entities.UserID, appID uuid.UUID) (*entities.App, error) {
	ctx, span, ctxLogger := service.tracer.StartWithLogger(ctx, service.logger)
	defer span.End()

	app, err := service.repository.Load(ctx, userID, appID)
	if err != nil {
		msg := fmt.Sprintf("cannot load [%T] with ID [%s] for user [%s]", &entities.App{}, appID, userID)
		return nil, service.tracer.WrapErrorSpan(span, stacktrace.Propagate(err, msg))
	}

	ctxLogger.Info(fmt.Sprintf("loaded [%T] with ID [%s] for user [%s]", app, appID, userID))
	return app, nil
}

// Delete an entities.App
func (service *AppService) Delete(ctx context.Context, userID entities.UserID, appID uuid.UUID) error {
	ctx, span, ctxLogger := service.tracer.StartWithLogger(ctx, service.logger)
	defer span.End()

	app, err := service.repository.Load(ctx, userID, appID)
	if err != nil {
		msg := fmt.Sprintf("cannot load [%T] with ID [%s] for user [%s]", &entities.App{}, appID, userID)
		return stacktrace.Propagate(err, msg)
	}

	if err = service.repository.Delete(ctx, app); err != nil {
		msg := fmt.Sprintf("cannot delete [%T] with ID [%s] for user [%s]", app, app.ID, userID)
		return service.tracer.WrapErrorSpan(span, stacktrace.Propagate(err, msg))
	}

	ctxLogger.Info(fmt.Sprintf("deleted [%T] with ID [%s] for user ID [%s]", app, app.ID, userID))
	return nil
}

// AddPhoneNumber adds a phone number to an entities.App
func (service *AppService) AddPhoneNumber(ctx context.Context, userID entities.UserID, appID uuid.UUID, phoneNumber string) error {
	ctx, span, ctxLogger := service.tracer.StartWithLogger(ctx, service.logger)
	defer span.End()

	app, err := service.repository.Load(ctx, userID, appID)
	if err != nil {
		msg := fmt.Sprintf("cannot load [%T] with ID [%s] for user [%s]", &entities.App{}, appID, userID)
		return stacktrace.Propagate(err, msg)
	}

	if err = service.repository.AddPhoneNumber(ctx, app, phoneNumber); err != nil {
		msg := fmt.Sprintf("cannot add phone number [%s] to [%T] with ID [%s]", phoneNumber, app, appID)
		return service.tracer.WrapErrorSpan(span, stacktrace.Propagate(err, msg))
	}

	ctxLogger.Info(fmt.Sprintf("added phone number [%s] to [%T] with ID [%s] for user [%s]", phoneNumber, app, appID, userID))
	return nil
}

// RemovePhoneNumber removes a phone number from an entities.App
func (service *AppService) RemovePhoneNumber(ctx context.Context, userID entities.UserID, appID uuid.UUID, phoneNumber string) error {
	ctx, span, ctxLogger := service.tracer.StartWithLogger(ctx, service.logger)
	defer span.End()

	app, err := service.repository.Load(ctx, userID, appID)
	if err != nil {
		msg := fmt.Sprintf("cannot load [%T] with ID [%s] for user [%s]", &entities.App{}, appID, userID)
		return stacktrace.Propagate(err, msg)
	}

	if err = service.repository.RemovePhoneNumber(ctx, app, phoneNumber); err != nil {
		msg := fmt.Sprintf("cannot remove phone number [%s] from [%T] with ID [%s]", phoneNumber, app, appID)
		return service.tracer.WrapErrorSpan(span, stacktrace.Propagate(err, msg))
	}

	ctxLogger.Info(fmt.Sprintf("removed phone number [%s] from [%T] with ID [%s] for user [%s]", phoneNumber, app, appID, userID))
	return nil
}

// DeleteAllForUser deletes all entities.App for a user
func (service *AppService) DeleteAllForUser(ctx context.Context, userID entities.UserID) error {
	ctx, span, ctxLogger := service.tracer.StartWithLogger(ctx, service.logger)
	defer span.End()

	if err := service.repository.DeleteAllForUser(ctx, userID); err != nil {
		msg := fmt.Sprintf("cannot delete all [%T] for user ID [%s]", &entities.App{}, userID)
		return stacktrace.Propagate(err, msg)
	}

	ctxLogger.Info(fmt.Sprintf("deleted all [%T] for user ID [%s]", &entities.App{}, userID))
	return nil
}

func (service *AppService) generateKey(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", stacktrace.Propagate(err, fmt.Sprintf("cannot generate [%d] random bytes", n))
	}
	return base64.URLEncoding.EncodeToString(b)[0:n], nil
}
