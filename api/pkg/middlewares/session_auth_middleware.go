package middlewares

import (
	"fmt"

	"github.com/NdoleStudio/httpsms/pkg/entities"
	"github.com/NdoleStudio/httpsms/pkg/repositories"
	"github.com/NdoleStudio/httpsms/pkg/telemetry"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/palantir/stacktrace"
)

// SessionCookieName is the name of the session cookie
const SessionCookieName = "httpsms_session"

// SessionAuth authenticates a user based on a session cookie stored in the database
func SessionAuth(logger telemetry.Logger, tracer telemetry.Tracer, repository repositories.SessionRepository) fiber.Handler {
	logger = logger.WithService("middlewares.SessionAuth")

	return func(c *fiber.Ctx) error {
		ctx, span, ctxLogger := tracer.StartFromFiberCtxWithLogger(c, logger, "middlewares.SessionAuth")
		defer span.End()

		sessionID := c.Cookies(SessionCookieName)
		if sessionID == "" {
			span.AddEvent("no session cookie found")
			return c.Next()
		}

		id, err := uuid.Parse(sessionID)
		if err != nil {
			ctxLogger.Warn(stacktrace.NewError(fmt.Sprintf("cannot parse session cookie value as UUID [%s]", sessionID)))
			c.ClearCookie(SessionCookieName)
			return c.Next()
		}

		session, err := repository.FindByID(ctx, id)
		if err != nil {
			if stacktrace.GetCode(err) == repositories.ErrCodeNotFound {
				c.ClearCookie(SessionCookieName)
				return c.Next()
			}
			ctxLogger.Error(stacktrace.Propagate(err, fmt.Sprintf("cannot find session with ID [%s]", id)))
			return c.Next()
		}

		if session.IsExpired() {
			_ = repository.DeleteByID(ctx, id)
			c.ClearCookie(SessionCookieName)
			return c.Next()
		}

		span.AddEvent(fmt.Sprintf("session [%s] is valid for user [%s]", id, session.UserID))
		c.Locals(ContextKeyAuthUserID, entities.AuthContext{
			ID:    session.UserID,
			Email: session.Email,
		})
		c.Locals(ContextKeyAuthSource, AuthSourceSession)
		return c.Next()
	}
}
