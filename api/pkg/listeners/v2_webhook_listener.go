package listeners

import (
	"context"
	"fmt"

	"github.com/NdoleStudio/httpsms/pkg/events"
	"github.com/NdoleStudio/httpsms/pkg/services"
	"github.com/NdoleStudio/httpsms/pkg/telemetry"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/palantir/stacktrace"
)

// V2WebhookListener sends v2-style delivery callbacks on message events
type V2WebhookListener struct {
	logger  telemetry.Logger
	tracer  telemetry.Tracer
	service *services.V2WebhookService
}

// NewV2WebhookListener creates a new V2WebhookListener
func NewV2WebhookListener(
	logger telemetry.Logger,
	tracer telemetry.Tracer,
	service *services.V2WebhookService,
) (l *V2WebhookListener, routes map[string]events.EventListener) {
	l = &V2WebhookListener{
		logger:  logger.WithService(fmt.Sprintf("%T", l)),
		tracer:  tracer,
		service: service,
	}

	return l, map[string]events.EventListener{
		events.EventTypeMessagePhoneDelivered: l.OnMessagePhoneDelivered,
		events.EventTypeMessageSendFailed:     l.OnMessageSendFailed,
		events.EventTypeMessagePhoneSent:      l.OnMessagePhoneSent,
	}
}

// OnMessagePhoneDelivered handles the message.phone.delivered event
func (l *V2WebhookListener) OnMessagePhoneDelivered(ctx context.Context, event cloudevents.Event) error {
	ctx, span := l.tracer.Start(ctx)
	defer span.End()

	var payload events.MessagePhoneDeliveredPayload
	if err := event.DataAs(&payload); err != nil {
		msg := fmt.Sprintf("cannot decode [%s] into [%T]", event.Data(), payload)
		return l.tracer.WrapErrorSpan(span, stacktrace.Propagate(err, msg))
	}

	messageID := fmt.Sprintf("sms_%s", payload.ID)
	l.service.SendDeliveryCallback(ctx, payload.UserID, payload.AppID, messageID, payload.Contact, services.V2DeliveryStatusDelivered, nil)
	return nil
}

// OnMessageSendFailed handles the message.send.failed event
func (l *V2WebhookListener) OnMessageSendFailed(ctx context.Context, event cloudevents.Event) error {
	ctx, span := l.tracer.Start(ctx)
	defer span.End()

	var payload events.MessageSendFailedPayload
	if err := event.DataAs(&payload); err != nil {
		msg := fmt.Sprintf("cannot decode [%s] into [%T]", event.Data(), payload)
		return l.tracer.WrapErrorSpan(span, stacktrace.Propagate(err, msg))
	}

	messageID := fmt.Sprintf("sms_%s", payload.ID)
	errCode := "MESSAGE_SEND_FAILED"
	l.service.SendDeliveryCallback(ctx, payload.UserID, payload.AppID, messageID, payload.Contact, services.V2DeliveryStatusFailed, &errCode)
	return nil
}

// OnMessagePhoneSent handles the message.phone.sent event
func (l *V2WebhookListener) OnMessagePhoneSent(ctx context.Context, event cloudevents.Event) error {
	ctx, span := l.tracer.Start(ctx)
	defer span.End()

	var payload events.MessagePhoneSentPayload
	if err := event.DataAs(&payload); err != nil {
		msg := fmt.Sprintf("cannot decode [%s] into [%T]", event.Data(), payload)
		return l.tracer.WrapErrorSpan(span, stacktrace.Propagate(err, msg))
	}

	messageID := fmt.Sprintf("sms_%s", payload.ID)
	l.service.SendDeliveryCallback(ctx, payload.UserID, payload.AppID, messageID, payload.Contact, services.V2DeliveryStatusSent, nil)
	return nil
}
