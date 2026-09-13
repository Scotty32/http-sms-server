package handlers

import (
	"fmt"
	"strings"

	"github.com/NdoleStudio/httpsms/pkg/entities"
	"github.com/NdoleStudio/httpsms/pkg/middlewares"
	"github.com/NdoleStudio/httpsms/pkg/requests"
	"github.com/NdoleStudio/httpsms/pkg/services"
	"github.com/NdoleStudio/httpsms/pkg/telemetry"
	"github.com/NdoleStudio/httpsms/pkg/validators"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/palantir/stacktrace"
)

// V2MessageHandler handles v2 message http requests, authenticated via an entities.App
// (ak_/x-api-secret) or an entities.PhoneAPIKey (pk_, x-api-key only).
// It mirrors AppHandler's send/status logic so external integrations (e.g. the OMCI SMS/WhatsApp
// provider contract) get a stable /v2 surface without duplicating validation/business logic.
type V2MessageHandler struct {
	handler
	logger         telemetry.Logger
	tracer         telemetry.Tracer
	validator      *validators.AppHandlerValidator
	messageService *services.MessageService
	billingService *services.BillingService
}

// NewV2MessageHandler creates a new V2MessageHandler
func NewV2MessageHandler(
	logger telemetry.Logger,
	tracer telemetry.Tracer,
	validator *validators.AppHandlerValidator,
	messageService *services.MessageService,
	billingService *services.BillingService,
) (h *V2MessageHandler) {
	return &V2MessageHandler{
		logger:         logger.WithService(fmt.Sprintf("%T", h)),
		tracer:         tracer,
		validator:      validator,
		messageService: messageService,
		billingService: billingService,
	}
}

// RegisterRoutes registers the v2 routes
func (h *V2MessageHandler) RegisterRoutes(router fiber.Router, middlewares ...fiber.Handler) {
	router.Post("/v2/send", h.computeRoute(middlewares, h.PostSend)...)
	router.Get("/v2/messages/:messageID/status", h.computeRoute(middlewares, h.MessageStatus)...)
}

// v2Error responds with the {"error": "CODE"} envelope, using the exact SmsErrorCode
// vocabulary expected by the calling provider-integration contract.
func (h *V2MessageHandler) v2Error(c *fiber.Ctx, status int, code string) error {
	return c.Status(status).JSON(fiber.Map{"error": code})
}

func (h *V2MessageHandler) authContext(c *fiber.Ctx) (entities.AuthContext, bool) {
	authCtx, ok := c.Locals(middlewares.ContextKeyAuthUserID).(entities.AuthContext)
	if !ok || authCtx.IsNoop() {
		return entities.AuthContext{}, false
	}
	return authCtx, true
}

// PostSend handles POST /v2/send
func (h *V2MessageHandler) PostSend(c *fiber.Ctx) error {
	ctx, span, ctxLogger := h.tracer.StartFromFiberCtxWithLogger(c, h.logger)
	defer span.End()

	authCtx, ok := h.authContext(c)
	if !ok {
		return h.v2Error(c, fiber.StatusUnauthorized, "API_KEY_INVALID")
	}

	if len(authCtx.PhoneNumbers) == 0 {
		ctxLogger.Warn(stacktrace.NewError(fmt.Sprintf("app for user [%s] has no phone numbers configured", authCtx.ID)))
		return h.v2Error(c, fiber.StatusForbidden, "FORBIDDEN")
	}

	var request requests.AppSendMessageRequest
	if err := c.BodyParser(&request); err != nil {
		ctxLogger.Warn(stacktrace.Propagate(err, fmt.Sprintf("cannot marshall [%s] into %T", c.Body(), request)))
		return h.v2Error(c, fiber.StatusBadRequest, "INTERNAL_ERROR")
	}

	if errors := h.validator.ValidateSendMessage(ctx, request.Sanitize()); len(errors) != 0 {
		return h.v2Error(c, fiber.StatusUnprocessableEntity, v2ValidationErrorCode(errors))
	}

	if msg := h.billingService.IsEntitled(ctx, authCtx.ID); msg != nil {
		ctxLogger.Warn(stacktrace.NewError(fmt.Sprintf("user [%s] not entitled to send message", authCtx.ID)))
		return h.v2Error(c, fiber.StatusPaymentRequired, "INTERNAL_ERROR")
	}

	from := authCtx.PhoneNumbers[0]
	message, err := h.messageService.SendMessage(ctx, request.ToMessageSendParams(authCtx.ID, authCtx.AppID, from, c.OriginalURL()))
	if err != nil {
		ctxLogger.Error(stacktrace.Propagate(err, fmt.Sprintf("cannot send v2 message for user [%s]", authCtx.ID)))
		return h.v2Error(c, fiber.StatusBadGateway, "MESSAGE_SEND_FAILED")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"messageId": fmt.Sprintf("sms_%s", message.ID),
		"status":    "queued",
	})
}

// MessageStatus handles GET /v2/messages/:messageID/status
func (h *V2MessageHandler) MessageStatus(c *fiber.Ctx) error {
	ctx, span, ctxLogger := h.tracer.StartFromFiberCtxWithLogger(c, h.logger)
	defer span.End()

	authCtx, ok := h.authContext(c)
	if !ok {
		return h.v2Error(c, fiber.StatusUnauthorized, "API_KEY_INVALID")
	}

	rawID := strings.TrimPrefix(c.Params("messageID"), "sms_")
	messageID, err := uuid.Parse(rawID)
	if err != nil {
		return h.v2Error(c, fiber.StatusNotFound, "NOT_FOUND")
	}

	message, err := h.messageService.LoadMessage(ctx, authCtx.ID, messageID)
	if err != nil {
		ctxLogger.Warn(stacktrace.Propagate(err, fmt.Sprintf("cannot load message [%s] for user [%s]", messageID, authCtx.ID)))
		return h.v2Error(c, fiber.StatusNotFound, "NOT_FOUND")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"messageId":  fmt.Sprintf("sms_%s", message.ID),
		"status":     mapMessageStatus(message.Status),
		"error_code": message.FailureReason,
	})
}

// v2ValidationErrorCode maps AppHandlerValidator field errors to the SmsErrorCode vocabulary
func v2ValidationErrorCode(errors map[string][]string) string {
	if _, ok := errors["to"]; ok {
		return "INVALID_PHONE_NUMBER"
	}
	if _, ok := errors["content"]; ok {
		return "EMPTY_MESSAGE"
	}
	return "INTERNAL_ERROR"
}
