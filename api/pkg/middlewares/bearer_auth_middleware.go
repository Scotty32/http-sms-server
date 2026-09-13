package middlewares

import (
	"fmt"
	"strings"

	"github.com/NdoleStudio/httpsms/pkg/entities"
	"github.com/NdoleStudio/httpsms/pkg/telemetry"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/palantir/stacktrace"
)

// BearerAuth authenticates a user based on a JWT bearer token
func BearerAuth(logger telemetry.Logger, tracer telemetry.Tracer, jwtSecret string) fiber.Handler {
	logger = logger.WithService("middlewares.BearerAuth")
	return func(c *fiber.Ctx) error {
		_, span := tracer.StartFromFiberCtx(c, "middlewares.BearerAuth")
		defer span.End()

		authToken := c.Get(authHeaderBearer)
		if len(authToken) <= len(bearerScheme) || !strings.HasPrefix(authToken, bearerScheme) {
			span.AddEvent(fmt.Sprintf("The request header has no [%s] token", bearerScheme))
			return c.Next()
		}

		tokenStr := strings.TrimSpace(authToken[len(bearerScheme):])
		if len(tokenStr) == 0 {
			return c.Next()
		}

		ctxLogger := tracer.CtxLogger(logger, span)

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, stacktrace.NewError(fmt.Sprintf("unexpected signing method: %v", token.Header["alg"]))
			}
			return []byte(jwtSecret), nil
		})
		if err != nil || !token.Valid {
			ctxLogger.Warn(stacktrace.Propagate(err, fmt.Sprintf("invalid JWT token [%s]", tokenStr)))
			return c.Next()
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			ctxLogger.Warn(stacktrace.NewError("cannot parse JWT claims"))
			return c.Next()
		}

		userID, _ := claims["user_id"].(string)
		email, _ := claims["email"].(string)
		if userID == "" || email == "" {
			ctxLogger.Warn(stacktrace.NewError("JWT claims missing user_id or email"))
			return c.Next()
		}

		span.AddEvent(fmt.Sprintf("[%s] JWT token is valid for user [%s]", bearerScheme, userID))
		c.Locals(ContextKeyAuthUserID, entities.AuthContext{
			ID:    entities.UserID(userID),
			Email: email,
		})
		c.Locals(ContextKeyAuthSource, AuthSourceBearer)
		return c.Next()
	}
}
