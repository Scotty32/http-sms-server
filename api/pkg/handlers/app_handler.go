package handlers

import (
	"fmt"
	"strings"

	"github.com/NdoleStudio/httpsms/pkg/entities"
	"github.com/NdoleStudio/httpsms/pkg/requests"
	"github.com/NdoleStudio/httpsms/pkg/services"
	"github.com/NdoleStudio/httpsms/pkg/telemetry"
	"github.com/NdoleStudio/httpsms/pkg/validators"
	"github.com/davecgh/go-spew/spew"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/palantir/stacktrace"
)

// AppHandler handles App http requests
type AppHandler struct {
	handler
	logger         telemetry.Logger
	tracer         telemetry.Tracer
	validator      *validators.AppHandlerValidator
	service        *services.AppService
	messageService *services.MessageService
	billingService *services.BillingService
}

// NewAppHandler creates a new AppHandler
func NewAppHandler(
	logger telemetry.Logger,
	tracer telemetry.Tracer,
	validator *validators.AppHandlerValidator,
	service *services.AppService,
	messageService *services.MessageService,
	billingService *services.BillingService,
) *AppHandler {
	return &AppHandler{
		logger:         logger.WithService(fmt.Sprintf("%T", &AppHandler{})),
		tracer:         tracer,
		validator:      validator,
		service:        service,
		messageService: messageService,
		billingService: billingService,
	}
}

// RegisterRoutes registers the routes for the AppHandler (requires user auth)
func (h *AppHandler) RegisterRoutes(app *fiber.App, middlewares ...fiber.Handler) {
	router := app.Group("/v1/apps")
	router.Get("/", h.computeRoute(middlewares, h.index)...)
	router.Post("/", h.computeRoute(middlewares, h.store)...)
	router.Get("/:appID", h.computeRoute(middlewares, h.show)...)
	router.Delete("/:appID", h.computeRoute(middlewares, h.delete)...)
	router.Post("/:appID/phone-numbers", h.computeRoute(middlewares, h.addPhoneNumber)...)
	router.Delete("/:appID/phone-numbers", h.computeRoute(middlewares, h.removePhoneNumber)...)
}

// RegisterAppKeyRoutes registers the routes that use App API key authentication
func (h *AppHandler) RegisterAppKeyRoutes(app *fiber.App, middlewares ...fiber.Handler) {
	router := app.Group("/v1/apps")
	router.Post("/messages/send", h.computeRoute(middlewares, h.sendMessage)...)
	router.Get("/messages/:messageID/status", h.computeRoute(middlewares, h.messageStatus)...)
}

// index lists all Apps for the authenticated user
func (h *AppHandler) index(c *fiber.Ctx) error {
	ctx, span, ctxLogger := h.tracer.StartFromFiberCtxWithLogger(c, h.logger)
	defer span.End()

	var request requests.AppIndex
	if err := c.QueryParser(&request); err != nil {
		msg := fmt.Sprintf("cannot marshall params [%s] into %T", c.OriginalURL(), request)
		ctxLogger.Warn(stacktrace.Propagate(err, msg))
		return h.responseBadRequest(c, err)
	}

	if errors := h.validator.ValidateIndex(ctx, request.Sanitize()); len(errors) != 0 {
		msg := fmt.Sprintf("validation errors [%s], while fetching apps [%+#v]", spew.Sdump(errors), request)
		ctxLogger.Warn(stacktrace.NewError(msg))
		return h.responseUnprocessableEntity(c, errors, "validation errors while fetching apps")
	}

	apps, err := h.service.Index(ctx, h.userIDFomContext(c), request.ToIndexParams())
	if err != nil {
		msg := fmt.Sprintf("cannot index apps with params [%+#v]", request)
		ctxLogger.Error(stacktrace.Propagate(err, msg))
		return h.responseInternalServerError(c)
	}

	return h.responseOK(c, fmt.Sprintf("fetched %d %s", len(apps), h.pluralize("app", len(apps))), apps)
}

// store creates a new App
func (h *AppHandler) store(c *fiber.Ctx) error {
	ctx, span, ctxLogger := h.tracer.StartFromFiberCtxWithLogger(c, h.logger)
	defer span.End()

	var request requests.AppStoreRequest
	if err := c.BodyParser(&request); err != nil {
		msg := fmt.Sprintf("cannot marshall [%s] into %T", c.Body(), request)
		ctxLogger.Warn(stacktrace.Propagate(err, msg))
		return h.responseBadRequest(c, err)
	}

	if errors := h.validator.ValidateStore(ctx, request.Sanitize()); len(errors) != 0 {
		msg := fmt.Sprintf("validation errors [%s], while storing app [%s]", spew.Sdump(errors), c.Body())
		ctxLogger.Warn(stacktrace.NewError(msg))
		return h.responseUnprocessableEntity(c, errors, "validation errors while creating app")
	}

	app, err := h.service.Create(ctx, h.userFromContext(c), services.AppStoreParams{
		Name:          request.Name,
		WebhookURL:    request.WebhookURL,
		WebhookSecret: request.WebhookSecret,
	})
	if err != nil {
		msg := fmt.Sprintf("cannot create app with payload [%s]", c.Body())
		ctxLogger.Error(stacktrace.Propagate(err, msg))
		return h.responseInternalServerError(c)
	}

	return h.responseCreated(c, "app created successfully", app)
}

// show returns a single App
func (h *AppHandler) show(c *fiber.Ctx) error {
	ctx, span, ctxLogger := h.tracer.StartFromFiberCtxWithLogger(c, h.logger)
	defer span.End()

	appID, err := uuid.Parse(c.Params("appID"))
	if err != nil {
		msg := fmt.Sprintf("cannot parse appID [%s] as uuid", c.Params("appID"))
		ctxLogger.Warn(stacktrace.Propagate(err, msg))
		return h.responseBadRequest(c, err)
	}

	app, err := h.service.Load(ctx, h.userIDFomContext(c), appID)
	if err != nil {
		msg := fmt.Sprintf("cannot load app with ID [%s]", appID)
		ctxLogger.Error(stacktrace.Propagate(err, msg))
		return h.responseNotFound(c, msg)
	}

	return h.responseOK(c, "app fetched successfully", app)
}

// delete removes an App
func (h *AppHandler) delete(c *fiber.Ctx) error {
	ctx, span, ctxLogger := h.tracer.StartFromFiberCtxWithLogger(c, h.logger)
	defer span.End()

	appID, err := uuid.Parse(c.Params("appID"))
	if err != nil {
		msg := fmt.Sprintf("cannot parse appID [%s] as uuid", c.Params("appID"))
		ctxLogger.Warn(stacktrace.Propagate(err, msg))
		return h.responseBadRequest(c, err)
	}

	if err = h.service.Delete(ctx, h.userIDFomContext(c), appID); err != nil {
		msg := fmt.Sprintf("cannot delete app with ID [%s]", appID)
		ctxLogger.Error(stacktrace.Propagate(err, msg))
		return h.responseInternalServerError(c)
	}

	return h.responseNoContent(c, "app deleted successfully")
}

// addPhoneNumber adds a phone number to an App
func (h *AppHandler) addPhoneNumber(c *fiber.Ctx) error {
	ctx, span, ctxLogger := h.tracer.StartFromFiberCtxWithLogger(c, h.logger)
	defer span.End()

	appID, err := uuid.Parse(c.Params("appID"))
	if err != nil {
		msg := fmt.Sprintf("cannot parse appID [%s] as uuid", c.Params("appID"))
		ctxLogger.Warn(stacktrace.Propagate(err, msg))
		return h.responseBadRequest(c, err)
	}

	var request requests.AppPhoneNumberRequest
	if err = c.BodyParser(&request); err != nil {
		msg := fmt.Sprintf("cannot marshall [%s] into %T", c.Body(), request)
		ctxLogger.Warn(stacktrace.Propagate(err, msg))
		return h.responseBadRequest(c, err)
	}

	if errors := h.validator.ValidatePhoneNumber(ctx, request.Sanitize()); len(errors) != 0 {
		msg := fmt.Sprintf("validation errors [%s], while adding phone number [%s]", spew.Sdump(errors), c.Body())
		ctxLogger.Warn(stacktrace.NewError(msg))
		return h.responseUnprocessableEntity(c, errors, "validation errors while adding phone number")
	}

	if err = h.service.AddPhoneNumber(ctx, h.userIDFomContext(c), appID, request.PhoneNumber); err != nil {
		msg := fmt.Sprintf("cannot add phone number [%s] to app [%s]", request.PhoneNumber, appID)
		ctxLogger.Error(stacktrace.Propagate(err, msg))
		return h.responseInternalServerError(c)
	}

	return h.responseOK(c, "phone number added to app", nil)
}

// removePhoneNumber removes a phone number from an App
func (h *AppHandler) removePhoneNumber(c *fiber.Ctx) error {
	ctx, span, ctxLogger := h.tracer.StartFromFiberCtxWithLogger(c, h.logger)
	defer span.End()

	appID, err := uuid.Parse(c.Params("appID"))
	if err != nil {
		msg := fmt.Sprintf("cannot parse appID [%s] as uuid", c.Params("appID"))
		ctxLogger.Warn(stacktrace.Propagate(err, msg))
		return h.responseBadRequest(c, err)
	}

	var request requests.AppPhoneNumberRequest
	if err = c.BodyParser(&request); err != nil {
		msg := fmt.Sprintf("cannot marshall [%s] into %T", c.Body(), request)
		ctxLogger.Warn(stacktrace.Propagate(err, msg))
		return h.responseBadRequest(c, err)
	}

	if errors := h.validator.ValidatePhoneNumber(ctx, request.Sanitize()); len(errors) != 0 {
		msg := fmt.Sprintf("validation errors [%s], while removing phone number [%s]", spew.Sdump(errors), c.Body())
		ctxLogger.Warn(stacktrace.NewError(msg))
		return h.responseUnprocessableEntity(c, errors, "validation errors while removing phone number")
	}

	if err = h.service.RemovePhoneNumber(ctx, h.userIDFomContext(c), appID, request.PhoneNumber); err != nil {
		msg := fmt.Sprintf("cannot remove phone number [%s] from app [%s]", request.PhoneNumber, appID)
		ctxLogger.Error(stacktrace.Propagate(err, msg))
		return h.responseInternalServerError(c)
	}

	return h.responseOK(c, "phone number removed from app", nil)
}

// sendMessage sends an SMS message via the App's phone numbers
func (h *AppHandler) sendMessage(c *fiber.Ctx) error {
	ctx, span, ctxLogger := h.tracer.StartFromFiberCtxWithLogger(c, h.logger)
	defer span.End()

	authCtx := h.userFromContext(c)

	if len(authCtx.PhoneNumbers) == 0 {
		ctxLogger.Warn(stacktrace.NewError(fmt.Sprintf("app for user [%s] has no phone numbers configured", authCtx.ID)))
		return h.responseUnprocessableEntity(c, nil, "this app has no phone numbers configured. Add at least one phone number to send messages")
	}

	var request requests.AppSendMessageRequest
	if err := c.BodyParser(&request); err != nil {
		msg := fmt.Sprintf("cannot marshall [%s] into %T", c.Body(), request)
		ctxLogger.Warn(stacktrace.Propagate(err, msg))
		return h.responseBadRequest(c, err)
	}

	if errors := h.validator.ValidateSendMessage(ctx, request.Sanitize()); len(errors) != 0 {
		msg := fmt.Sprintf("validation errors [%s], while sending message via app [%s]", spew.Sdump(errors), c.Body())
		ctxLogger.Warn(stacktrace.NewError(msg))
		return h.responseUnprocessableEntity(c, errors, "validation errors while sending message")
	}

	if msg := h.billingService.IsEntitled(ctx, authCtx.ID); msg != nil {
		ctxLogger.Warn(stacktrace.NewError(fmt.Sprintf("user with ID [%s] can't send a message", authCtx.ID)))
		return h.responsePaymentRequired(c, *msg)
	}

	// Pick the first phone number from the app's phone numbers
	from := authCtx.PhoneNumbers[0]

	message, err := h.messageService.SendMessage(ctx, request.ToMessageSendParams(authCtx.ID, authCtx.AppID, from, c.OriginalURL()))
	if err != nil {
		msg := fmt.Sprintf("cannot send message with payload [%s]", c.Body())
		ctxLogger.Error(stacktrace.Propagate(err, msg))
		return h.responseInternalServerError(c)
	}

	return h.responseOK(c, "message added to queue", message)
}

// messageStatus returns the delivery status of a message
func (h *AppHandler) messageStatus(c *fiber.Ctx) error {
	ctx, span, ctxLogger := h.tracer.StartFromFiberCtxWithLogger(c, h.logger)
	defer span.End()

	authCtx := h.userFromContext(c)

	rawID := strings.TrimPrefix(c.Params("messageID"), "sms_")
	messageID, err := uuid.Parse(rawID)
	if err != nil {
		ctxLogger.Warn(stacktrace.Propagate(err, fmt.Sprintf("cannot parse message ID [%s]", c.Params("messageID"))))
		return h.responseBadRequest(c, err)
	}

	message, err := h.messageService.LoadMessage(ctx, authCtx.ID, messageID)
	if err != nil {
		ctxLogger.Warn(stacktrace.Propagate(err, fmt.Sprintf("cannot load message [%s] for user [%s]", messageID, authCtx.ID)))
		return h.responseNotFound(c, "message not found")
	}

	return h.responseOK(c, "message status", fiber.Map{
		"status":     mapMessageStatus(message.Status),
		"error_code": message.FailureReason,
	})
}

// mapMessageStatus maps internal message status to the provider contract vocabulary
func mapMessageStatus(status entities.MessageStatus) string {
	switch status {
	case entities.MessageStatusDelivered:
		return "delivered"
	case entities.MessageStatusSent:
		return "sent"
	case entities.MessageStatusFailed, entities.MessageStatusExpired:
		return "failed"
	default:
		return "pending"
	}
}
