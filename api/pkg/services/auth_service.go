package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/NdoleStudio/httpsms/pkg/entities"
	"github.com/NdoleStudio/httpsms/pkg/repositories"
	"github.com/NdoleStudio/httpsms/pkg/telemetry"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/palantir/stacktrace"
	"golang.org/x/crypto/bcrypt"
)

// AuthService handles internal authentication (register / login)
type AuthService struct {
	service
	logger            telemetry.Logger
	tracer            telemetry.Tracer
	jwtSecret         string
	sessionDuration   time.Duration
	repository        repositories.UserRepository
	sessionRepository repositories.SessionRepository
	dispatcher        *EventDispatcher
}

// NewAuthService creates a new AuthService
func NewAuthService(
	logger telemetry.Logger,
	tracer telemetry.Tracer,
	jwtSecret string,
	sessionDuration time.Duration,
	repository repositories.UserRepository,
	sessionRepository repositories.SessionRepository,
	dispatcher *EventDispatcher,
) *AuthService {
	return &AuthService{
		logger:            logger.WithService(fmt.Sprintf("%T", &AuthService{})),
		tracer:            tracer,
		jwtSecret:         jwtSecret,
		sessionDuration:   sessionDuration,
		repository:        repository,
		sessionRepository: sessionRepository,
		dispatcher:        dispatcher,
	}
}

// AuthResult is returned after successful registration or login
type AuthResult struct {
	User    *entities.User
	Session *entities.Session
}

// Register creates a new user with the given name, email and password
func (service *AuthService) Register(ctx context.Context, source string, name string, email string, password string) (*AuthResult, error) {
	ctx, span, ctxLogger := service.tracer.StartWithLogger(ctx, service.logger)
	defer span.End()

	// Check if email is already taken
	existing, err := service.repository.LoadByEmail(ctx, email)
	if err == nil && existing != nil {
		return nil, service.tracer.WrapErrorSpan(span, stacktrace.NewErrorWithCode(ErrCodeConflict, fmt.Sprintf("user with email [%s] already exists", email)))
	}

	if err != nil && stacktrace.GetCode(err) != repositories.ErrCodeNotFound {
		return nil, service.tracer.WrapErrorSpan(span, stacktrace.Propagate(err, fmt.Sprintf("cannot check email [%s]", email)))
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, service.tracer.WrapErrorSpan(span, stacktrace.Propagate(err, "cannot hash password"))
	}

	apiKey, err := service.generateKey(64)
	if err != nil {
		return nil, service.tracer.WrapErrorSpan(span, stacktrace.Propagate(err, "cannot generate API key"))
	}

	user := &entities.User{
		ID:               entities.UserID(uuid.New().String()),
		Name:             name,
		Email:            email,
		PasswordHash:     string(passwordHash),
		APIKey:           "uk_" + apiKey,
		SubscriptionName: entities.SubscriptionNameFree,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}

	if err = service.repository.Store(ctx, user); err != nil {
		return nil, service.tracer.WrapErrorSpan(span, stacktrace.Propagate(err, fmt.Sprintf("cannot create user with email [%s]", email)))
	}

	session, err := service.createSession(ctx, user)
	if err != nil {
		return nil, service.tracer.WrapErrorSpan(span, stacktrace.Propagate(err, "cannot create session"))
	}

	ctxLogger.Info(fmt.Sprintf("registered new user with ID [%s] and email [%s]", user.ID, user.Email))

	service.dispatchUserCreatedEvent(ctx, source, user)

	return &AuthResult{User: user, Session: session}, nil
}

// Login verifies the email+password and creates a session
func (service *AuthService) Login(ctx context.Context, email string, password string) (*AuthResult, error) {
	ctx, span, ctxLogger := service.tracer.StartWithLogger(ctx, service.logger)
	defer span.End()

	user, err := service.repository.LoadByEmail(ctx, strings.ToLower(email))
	if err != nil {
		if stacktrace.GetCode(err) == repositories.ErrCodeNotFound {
			return nil, service.tracer.WrapErrorSpan(span, stacktrace.NewErrorWithCode(ErrCodeUnauthorized, "invalid email or password"))
		}
		return nil, service.tracer.WrapErrorSpan(span, stacktrace.Propagate(err, fmt.Sprintf("cannot load user with email [%s]", email)))
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, service.tracer.WrapErrorSpan(span, stacktrace.NewErrorWithCode(ErrCodeUnauthorized, "invalid email or password"))
		}
		return nil, service.tracer.WrapErrorSpan(span, stacktrace.Propagate(err, "cannot compare password hash"))
	}

	session, err := service.createSession(ctx, user)
	if err != nil {
		return nil, service.tracer.WrapErrorSpan(span, stacktrace.Propagate(err, "cannot create session"))
	}

	ctxLogger.Info(fmt.Sprintf("user [%s] logged in successfully", user.ID))
	return &AuthResult{User: user, Session: session}, nil
}

// Logout deletes the session with the given ID
func (service *AuthService) Logout(ctx context.Context, sessionID uuid.UUID) error {
	ctx, span, ctxLogger := service.tracer.StartWithLogger(ctx, service.logger)
	defer span.End()

	if err := service.sessionRepository.DeleteByID(ctx, sessionID); err != nil {
		if stacktrace.GetCode(err) == repositories.ErrCodeNotFound {
			return nil
		}
		return service.tracer.WrapErrorSpan(span, stacktrace.Propagate(err, fmt.Sprintf("cannot delete session with ID [%s]", sessionID)))
	}

	ctxLogger.Info(fmt.Sprintf("deleted session with ID [%s]", sessionID))
	return nil
}

// GenerateToken creates a signed JWT for the given user (24h expiry) — kept for backward compat
func (service *AuthService) GenerateToken(user *entities.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"email":   user.Email,
		"exp":     time.Now().UTC().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().UTC().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(service.jwtSecret))
	if err != nil {
		return "", stacktrace.Propagate(err, "cannot sign JWT token")
	}
	return signed, nil
}

func (service *AuthService) createSession(ctx context.Context, user *entities.User) (*entities.Session, error) {
	ctx, span := service.tracer.Start(ctx)
	defer span.End()

	duration := service.sessionDuration
	if duration == 0 {
		duration = 30 * 24 * time.Hour
	}

	session := &entities.Session{
		ID:        uuid.New(),
		UserID:    user.ID,
		Email:     user.Email,
		ExpiresAt: time.Now().UTC().Add(duration),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := service.sessionRepository.Store(ctx, session); err != nil {
		return nil, service.tracer.WrapErrorSpan(span, stacktrace.Propagate(err, "cannot store session"))
	}

	return session, nil
}

func (service *AuthService) dispatchUserCreatedEvent(ctx context.Context, source string, user *entities.User) {
	ctx, span, ctxLogger := service.tracer.StartWithLogger(ctx, service.logger)
	defer span.End()

	event, err := service.createEvent("user.account.created", source, map[string]interface{}{
		"user_id":    user.ID,
		"user_email": user.Email,
		"timestamp":  time.Now().UTC(),
	})
	if err != nil {
		ctxLogger.Error(stacktrace.Propagate(err, fmt.Sprintf("cannot create event for user [%s]", user.ID)))
		return
	}
	if err = service.dispatcher.Dispatch(ctx, event); err != nil {
		ctxLogger.Error(stacktrace.Propagate(err, fmt.Sprintf("cannot dispatch event for user [%s]", user.ID)))
	}
}

func (service *AuthService) generateKey(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", stacktrace.Propagate(err, fmt.Sprintf("cannot generate [%d] random bytes", n))
	}
	return base64.URLEncoding.EncodeToString(b)[0:n], nil
}

// ErrCodeConflict is returned when a resource already exists
const ErrCodeConflict = stacktrace.ErrorCode(409)

// ErrCodeUnauthorized is returned when credentials are invalid
const ErrCodeUnauthorized = stacktrace.ErrorCode(401)
