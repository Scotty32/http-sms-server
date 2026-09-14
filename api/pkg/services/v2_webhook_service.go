package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/NdoleStudio/httpsms/pkg/entities"
	"github.com/NdoleStudio/httpsms/pkg/repositories"
	"github.com/NdoleStudio/httpsms/pkg/telemetry"
	"github.com/google/uuid"
	"github.com/palantir/stacktrace"
)

// ErrCodeWebhookTestFailed is returned when a test payload could not be delivered to a candidate webhook URL
const ErrCodeWebhookTestFailed = stacktrace.ErrorCode(4022)

// v2WebhookUserAgent identifies httpSMS webhook requests, since Go's default "Go-http-client" User-Agent
// gets flagged by some receivers' bot/WAF protection (e.g. Cloudflare)
const v2WebhookUserAgent = "httpSMS-webhook/1.0 (+https://httpsms.com)"

// V2DeliveryStatus represents the delivery status sent in the webhook callback
type V2DeliveryStatus string

const (
	V2DeliveryStatusDelivered V2DeliveryStatus = "delivered"
	V2DeliveryStatusFailed    V2DeliveryStatus = "failed"
	V2DeliveryStatusSent      V2DeliveryStatus = "sent"
)

// v2WebhookPayload is the body sent to the webhook URL
type v2WebhookPayload struct {
	MessageID string           `json:"message_id"`
	Status    V2DeliveryStatus `json:"status"`
	ErrorCode *string          `json:"error_code"`
	Recipient string           `json:"recipient"`
	Timestamp string           `json:"timestamp"`
}

// V2WebhookService sends delivery callbacks for v2 API users
type V2WebhookService struct {
	service
	logger         telemetry.Logger
	tracer         telemetry.Tracer
	client         *http.Client
	userRepository repositories.UserRepository
	appRepository  repositories.AppRepository
}

// NewV2WebhookService creates a new V2WebhookService
func NewV2WebhookService(
	logger telemetry.Logger,
	tracer telemetry.Tracer,
	client *http.Client,
	userRepository repositories.UserRepository,
	appRepository repositories.AppRepository,
) (s *V2WebhookService) {
	return &V2WebhookService{
		logger:         logger.WithService(fmt.Sprintf("%T", s)),
		tracer:         tracer,
		client:         client,
		userRepository: userRepository,
		appRepository:  appRepository,
	}
}

// SendDeliveryCallback sends the delivery callback to the app's or user's webhook_url.
func (s *V2WebhookService) SendDeliveryCallback(ctx context.Context, userID entities.UserID, appID *uuid.UUID, messageID, recipient string, status V2DeliveryStatus, errorCode *string) {
	ctx, span, ctxLogger := s.tracer.StartWithLogger(ctx, s.logger)
	defer span.End()

	var webhookURL string
	var webhookSecret string

	// Try app-level webhook first
	if appID != nil {
		app, err := s.appRepository.Load(ctx, userID, *appID)
		if err != nil {
			ctxLogger.Warn(stacktrace.Propagate(err, fmt.Sprintf("cannot load app [%s] for user [%s] for v2 delivery webhook", appID, userID)))
		} else if app.WebhookURL != nil && *app.WebhookURL != "" {
			webhookURL = *app.WebhookURL
			if app.WebhookSecret != nil {
				webhookSecret = *app.WebhookSecret
			}
		}
	}

	// Fall back to user-level webhook
	if webhookURL == "" {
		user, err := s.userRepository.Load(ctx, userID)
		if err != nil {
			ctxLogger.Warn(stacktrace.Propagate(err, fmt.Sprintf("cannot load user [%s] for v2 delivery webhook", userID)))
			return
		}

		if user.WebhookURL == nil || *user.WebhookURL == "" {
			ctxLogger.Info(fmt.Sprintf("user [%s] has no webhook_url configured, skipping v2 delivery callback", userID))
			return
		}
		webhookURL = *user.WebhookURL
		if user.WebhookSecret != nil {
			webhookSecret = *user.WebhookSecret
		}
	}

	payload := v2WebhookPayload{
		MessageID: messageID,
		Status:    status,
		ErrorCode: errorCode,
		Recipient: recipient,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	// Use a background context so retries aren't cancelled when the parent request ends
	go s.sendWithRetry(context.Background(), webhookURL, webhookSecret, payload)
}

var v2WebhookBackoffs = []time.Duration{1 * time.Second, 5 * time.Second, 30 * time.Second}

func (s *V2WebhookService) sendWithRetry(ctx context.Context, webhookURL string, webhookSecret string, payload v2WebhookPayload) {
	ctx, span, ctxLogger := s.tracer.StartWithLogger(ctx, s.logger)
	defer span.End()

	body, err := json.Marshal(payload)
	if err != nil {
		ctxLogger.Error(stacktrace.Propagate(err, fmt.Sprintf("cannot marshal v2 webhook payload for url [%s]", webhookURL)))
		return
	}

	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(v2WebhookBackoffs[attempt-1])
		}

		reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, webhookURL, bytes.NewReader(body))
		if err != nil {
			cancel()
			ctxLogger.Error(stacktrace.Propagate(err, fmt.Sprintf("cannot create v2 webhook request for url [%s]", webhookURL)))
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", v2WebhookUserAgent)
		if webhookSecret != "" {
			mac := hmac.New(sha256.New, []byte(webhookSecret))
			mac.Write(body)
			signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))
			req.Header.Set("x-webhook-signature", signature)
		}

		resp, err := s.client.Do(req)
		cancel()
		if err != nil {
			ctxLogger.Warn(stacktrace.Propagate(err, fmt.Sprintf("v2 webhook attempt [%d] failed for url [%s]", attempt+1, webhookURL)))
			continue
		}
		_ = resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			ctxLogger.Info(fmt.Sprintf("v2 webhook sent successfully to [%s] on attempt [%d] with status [%d]", webhookURL, attempt+1, resp.StatusCode))
			return
		}

		ctxLogger.Warn(stacktrace.NewError(fmt.Sprintf("v2 webhook attempt [%d] got non-2xx status [%d] from url [%s]", attempt+1, resp.StatusCode, webhookURL)))
	}

	ctxLogger.Error(stacktrace.NewError(fmt.Sprintf("v2 webhook failed after 3 attempts for url [%s]", webhookURL)))
}

// TestWebhookURL sends a {"type":"test"} payload to a candidate webhook URL so a user can confirm it works before saving it.
// This mirrors the convention used by receivers such as OMCI, which respond {"ok":true} to this payload without processing
// it as a real delivery event.
func (s *V2WebhookService) TestWebhookURL(ctx context.Context, webhookURL string, webhookSecret string) error {
	ctx, span, ctxLogger := s.tracer.StartWithLogger(ctx, s.logger)
	defer span.End()

	if err := validateWebhookURL(webhookURL); err != nil {
		ctxLogger.Warn(stacktrace.Propagate(err, fmt.Sprintf("invalid webhook url [%s]", webhookURL)))
		return stacktrace.NewErrorWithCode(ErrCodeWebhookTestFailed, err.Error())
	}

	body, err := json.Marshal(map[string]string{"type": "test"})
	if err != nil {
		return s.tracer.WrapErrorSpan(span, stacktrace.Propagate(err, "cannot marshal webhook test payload"))
	}

	reqCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, webhookURL, bytes.NewReader(body))
	if err != nil {
		return s.tracer.WrapErrorSpan(span, stacktrace.Propagate(err, fmt.Sprintf("cannot create test request for url [%s]", webhookURL)))
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", v2WebhookUserAgent)
	if webhookSecret != "" {
		mac := hmac.New(sha256.New, []byte(webhookSecret))
		mac.Write(body)
		req.Header.Set("x-webhook-signature", "sha256="+hex.EncodeToString(mac.Sum(nil)))
	}

	ctxLogger.Info(fmt.Sprintf("sending webhook test request [%s %s], headers %+v, body [%s]", req.Method, webhookURL, req.Header, string(body)))

	resp, err := s.client.Do(req)
	if err != nil {
		ctxLogger.Warn(stacktrace.Propagate(err, fmt.Sprintf("webhook test request failed for url [%s]", webhookURL)))
		return stacktrace.NewErrorWithCode(ErrCodeWebhookTestFailed, fmt.Sprintf("could not reach [%s]: %s", webhookURL, err.Error()))
	}
	defer func() { _ = resp.Body.Close() }()

	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		ctxLogger.Warn(stacktrace.Propagate(err, fmt.Sprintf("cannot read webhook test response body for url [%s]", webhookURL)))
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		ctxLogger.Warn(stacktrace.NewError(fmt.Sprintf("webhook test response [%s %s], status [%d], headers %+v, body [%s]", req.Method, webhookURL, resp.StatusCode, resp.Header, string(responseBody))))
		return stacktrace.NewErrorWithCode(ErrCodeWebhookTestFailed, fmt.Sprintf("webhook url [%s] responded with status [%d], body [%s]", webhookURL, resp.StatusCode, string(responseBody)))
	}

	ctxLogger.Info(fmt.Sprintf("webhook test succeeded for url [%s] with status [%d], body [%s]", webhookURL, resp.StatusCode, string(responseBody)))
	return nil
}

// validateWebhookURL rejects malformed URLs and, unless WEBHOOK_TEST_ALLOW_PRIVATE_IPS=true (used in local/dev
// environments where the webhook target may live on a private docker network), URLs resolving to a private,
// loopback, or link-local address, to prevent this endpoint from being used as an SSRF vector.
func validateWebhookURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("cannot parse webhook url: %w", err)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("webhook url scheme must be http or https")
	}

	if parsed.Hostname() == "" {
		return fmt.Errorf("webhook url must have a host")
	}

	if os.Getenv("WEBHOOK_TEST_ALLOW_PRIVATE_IPS") == "true" {
		return nil
	}

	ips, err := net.LookupIP(parsed.Hostname())
	if err != nil {
		return fmt.Errorf("cannot resolve webhook host [%s]: %w", parsed.Hostname(), err)
	}

	for _, ip := range ips {
		if isPrivateOrReservedIP(ip) {
			return fmt.Errorf("webhook url resolves to a private/reserved IP address [%s] which is not allowed", ip)
		}
	}

	return nil
}

func isPrivateOrReservedIP(ip net.IP) bool {
	return ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified()
}
