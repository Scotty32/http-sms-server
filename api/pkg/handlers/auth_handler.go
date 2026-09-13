package handlers

import (
	"fmt"
	"os"
	"time"

	"github.com/NdoleStudio/httpsms/pkg/middlewares"
	"github.com/NdoleStudio/httpsms/pkg/requests"
	"github.com/NdoleStudio/httpsms/pkg/services"
	"github.com/NdoleStudio/httpsms/pkg/telemetry"
	"github.com/NdoleStudio/httpsms/pkg/validators"
	"github.com/davecgh/go-spew/spew"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/palantir/stacktrace"
)

// AuthHandler handles internal auth http requests
type AuthHandler struct {
	handler
	logger    telemetry.Logger
	tracer    telemetry.Tracer
	validator *validators.AuthHandlerValidator
	service   *services.AuthService
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(
	logger telemetry.Logger,
	tracer telemetry.Tracer,
	validator *validators.AuthHandlerValidator,
	service *services.AuthService,
) *AuthHandler {
	return &AuthHandler{
		logger:    logger.WithService(fmt.Sprintf("%T", &AuthHandler{})),
		tracer:    tracer,
		validator: validator,
		service:   service,
	}
}

// RegisterRoutes registers auth routes (public, no auth required)
func (h *AuthHandler) RegisterRoutes(router fiber.Router) {
	router.Post("/v1/auth/register", h.register)
	router.Post("/v1/auth/login", h.login)
	router.Delete("/v1/auth/sessions", h.logout)
}

// register creates a new user account
func (h *AuthHandler) register(c *fiber.Ctx) error {
	ctx, span, ctxLogger := h.tracer.StartFromFiberCtxWithLogger(c, h.logger)
	defer span.End()

	var request requests.AuthRegisterRequest
	if err := c.BodyParser(&request); err != nil {
		msg := fmt.Sprintf("cannot marshall [%s] into %T", c.Body(), request)
		ctxLogger.Warn(stacktrace.Propagate(err, msg))
		return h.responseBadRequest(c, err)
	}

	if errors := h.validator.ValidateRegister(ctx, request.Sanitize()); len(errors) != 0 {
		msg := fmt.Sprintf("validation errors [%s], while registering user", spew.Sdump(errors))
		ctxLogger.Warn(stacktrace.NewError(msg))
		return h.responseUnprocessableEntity(c, errors, "validation errors while registering")
	}

	result, err := h.service.Register(ctx, c.OriginalURL(), request.Name, request.Email, request.Password)
	if err != nil {
		if stacktrace.GetCode(err) == services.ErrCodeConflict {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"status":  "error",
				"message": "An account with this email already exists",
			})
		}
		ctxLogger.Error(stacktrace.Propagate(err, "cannot register user"))
		return h.responseInternalServerError(c)
	}

	h.setSessionCookie(c, result.Session.ID, result.Session.ExpiresAt)

	token, err := h.service.GenerateToken(result.User)
	if err != nil {
		ctxLogger.Error(stacktrace.Propagate(err, "cannot generate JWT token"))
		return h.responseInternalServerError(c)
	}

	return h.responseCreated(c, "account created successfully", fiber.Map{
		"user":  result.User,
		"token": token,
	})
}

// login authenticates a user and sets a session cookie
func (h *AuthHandler) login(c *fiber.Ctx) error {
	ctx, span, ctxLogger := h.tracer.StartFromFiberCtxWithLogger(c, h.logger)
	defer span.End()

	var request requests.AuthLoginRequest
	if err := c.BodyParser(&request); err != nil {
		msg := fmt.Sprintf("cannot marshall [%s] into %T", c.Body(), request)
		ctxLogger.Warn(stacktrace.Propagate(err, msg))
		return h.responseBadRequest(c, err)
	}

	if errors := h.validator.ValidateLogin(ctx, request.Sanitize()); len(errors) != 0 {
		msg := fmt.Sprintf("validation errors [%s], while logging in", spew.Sdump(errors))
		ctxLogger.Warn(stacktrace.NewError(msg))
		return h.responseUnprocessableEntity(c, errors, "validation errors while logging in")
	}

	result, err := h.service.Login(ctx, request.Email, request.Password)
	if err != nil {
		if stacktrace.GetCode(err) == services.ErrCodeUnauthorized {
			return h.responseUnauthorized(c)
		}
		ctxLogger.Error(stacktrace.Propagate(err, "cannot login user"))
		return h.responseInternalServerError(c)
	}

	h.setSessionCookie(c, result.Session.ID, result.Session.ExpiresAt)

	token, err := h.service.GenerateToken(result.User)
	if err != nil {
		ctxLogger.Error(stacktrace.Propagate(err, "cannot generate JWT token"))
		return h.responseInternalServerError(c)
	}

	return h.responseOK(c, "login successful", fiber.Map{
		"user":  result.User,
		"token": token,
	})
}

// logout deletes the current session and clears the cookie
func (h *AuthHandler) logout(c *fiber.Ctx) error {
	ctx, span, ctxLogger := h.tracer.StartFromFiberCtxWithLogger(c, h.logger)
	defer span.End()

	sessionID := c.Cookies(middlewares.SessionCookieName)
	if sessionID != "" {
		id, err := uuid.Parse(sessionID)
		if err == nil {
			if err = h.service.Logout(ctx, id); err != nil {
				ctxLogger.Error(stacktrace.Propagate(err, fmt.Sprintf("cannot logout session [%s]", id)))
			}
		}
	}

	c.ClearCookie(middlewares.SessionCookieName)
	return h.responseNoContent(c, "logged out successfully")
}

func (h *AuthHandler) setSessionCookie(c *fiber.Ctx, sessionID uuid.UUID, expiresAt time.Time) {
	secure := c.Protocol() == "https" || os.Getenv("ENV") == "production"

	sameSite := "Lax"
	if secure {
		// Cross-origin requests (e.g. web app on a different domain than the API,
		// such as behind a Cloudflare tunnel) only send cookies when SameSite=None,
		// which browsers require to be paired with Secure.
		sameSite = "None"
	}

	c.Cookie(&fiber.Cookie{
		Name:     middlewares.SessionCookieName,
		Value:    sessionID.String(),
		Expires:  expiresAt,
		HTTPOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		Path:     "/",
	})
}
