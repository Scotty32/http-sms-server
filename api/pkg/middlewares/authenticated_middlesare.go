package middlewares

import (
	"github.com/NdoleStudio/httpsms/pkg/entities"
	"github.com/NdoleStudio/httpsms/pkg/telemetry"
	"github.com/gofiber/fiber/v2"
)

const (
	authHeaderBearer = "Authorization"
	authHeaderAPIKey = "x-api-key"
	bearerScheme     = "Bearer"
)

const (
	// ContextKeyAuthUserID is the context key used to store the ID of an authenticated user
	ContextKeyAuthUserID = "auth.user.id"

	// ContextKeyAuthSource is the context key used to store the authentication source
	ContextKeyAuthSource = "auth.source"

	// AuthSourceSession indicates authentication via session cookie
	AuthSourceSession = "session"
	// AuthSourceBearer indicates authentication via JWT bearer token
	AuthSourceBearer = "bearer"
	// AuthSourceAPIKey indicates authentication via x-api-key header (user or bearer API key)
	AuthSourceAPIKey = "api_key"
	// AuthSourcePhoneKey indicates authentication via phone API key (pk_)
	AuthSourcePhoneKey = "phone_api_key"
	// AuthSourceAppKey indicates authentication via app API key (ak_)
	AuthSourceAppKey = "app_key"
)

// Authenticated checks if the request is authenticated
func Authenticated(tracer telemetry.Tracer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		_, span := tracer.StartFromFiberCtx(c, "middlewares.Authenticated")
		defer span.End()

		if tokenUser, ok := c.Locals(ContextKeyAuthUserID).(entities.AuthContext); !ok || tokenUser.IsNoop() {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status":  "error",
				"message": "You are not authorized to carry out this request.",
				"data":    "Make sure your API key is set in the [x-api-key] header in the request",
			})
		}

		return c.Next()
	}
}

// SessionOnly rejects requests that were not authenticated via a session cookie
func SessionOnly(tracer telemetry.Tracer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		_, span := tracer.StartFromFiberCtx(c, "middlewares.SessionOnly")
		defer span.End()

		source, _ := c.Locals(ContextKeyAuthSource).(string)
		if source != AuthSourceSession {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status":  "error",
				"message": "This endpoint requires session authentication. Please log in via the web interface.",
			})
		}

		return c.Next()
	}
}
