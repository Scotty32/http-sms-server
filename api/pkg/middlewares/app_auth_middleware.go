package middlewares

import (
	"crypto/subtle"
	"fmt"
	"strings"

	"github.com/NdoleStudio/httpsms/pkg/entities"
	"github.com/NdoleStudio/httpsms/pkg/repositories"
	"github.com/NdoleStudio/httpsms/pkg/telemetry"
	"github.com/gofiber/fiber/v2"
	"github.com/palantir/stacktrace"
)

const authHeaderAPISecret = "x-api-secret"

// AppAuth authenticates a request using an App API key (ak_) and API secret (x-api-secret header)
func AppAuth(logger telemetry.Logger, tracer telemetry.Tracer, repository repositories.AppRepository) fiber.Handler {
	logger = logger.WithService("middlewares.AppAuth")

	return func(c *fiber.Ctx) error {
		ctx, span, ctxLogger := tracer.StartFromFiberCtxWithLogger(c, logger, "middlewares.AppAuth")
		defer span.End()

		apiKey := getAPIKeyFromRequest(c)
		if len(apiKey) == 0 || apiKey == "undefined" || !strings.HasPrefix(apiKey, "ak_") {
			span.AddEvent(fmt.Sprintf("the request header has no app api key [%s] header", authHeaderAPIKey))
			return c.Next()
		}

		apiSecret := c.Get(authHeaderAPISecret)
		if len(apiSecret) == 0 {
			ctxLogger.Warn(stacktrace.NewError(fmt.Sprintf("request with app api key [%s] has no [%s] header", apiKey, authHeaderAPISecret)))
			return c.Next()
		}

		app, err := repository.LoadByAPIKey(ctx, apiKey)
		if err != nil {
			ctxLogger.Error(stacktrace.Propagate(err, fmt.Sprintf("cannot load app with api key [%s]", apiKey)))
			return c.Next()
		}

		if subtle.ConstantTimeCompare([]byte(app.APISecret), []byte(apiSecret)) != 1 {
			ctxLogger.Warn(stacktrace.NewError(fmt.Sprintf("invalid api secret for app with api key [%s]", apiKey)))
			return c.Next()
		}

		authCtx := entities.AuthContext{
			ID:           app.UserID,
			Email:        app.UserEmail,
			AppID:        &app.ID,
			PhoneNumbers: app.PhoneNumbers,
		}

		c.Locals(ContextKeyAuthUserID, authCtx)
		c.Locals(ContextKeyAuthSource, AuthSourceAppKey)
		return c.Next()
	}
}
