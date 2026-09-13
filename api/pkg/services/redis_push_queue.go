package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/NdoleStudio/httpsms/pkg/telemetry"
	"github.com/carlmjohnson/requests"
	"github.com/hibiken/asynq"
	"github.com/palantir/stacktrace"
	"github.com/redis/go-redis/v9"
)

// RedisPushQueueTaskType identifies the asynq task type used by RedisPushQueue
const RedisPushQueueTaskType = "push_queue:http_task"

type redisPushQueue struct {
	config PushQueueConfig
	client *asynq.Client
	logger telemetry.Logger
	tracer telemetry.Tracer
}

// RedisPushQueue creates a PushQueue backed by Redis (via asynq): a self-hostable
// alternative to Google Cloud Tasks with automatic retries and persistence across restarts.
func RedisPushQueue(
	logger telemetry.Logger,
	tracer telemetry.Tracer,
	redisClient redis.UniversalClient,
	config PushQueueConfig,
) PushQueue {
	return &redisPushQueue{
		tracer: tracer,
		logger: logger.WithService(fmt.Sprintf("%T", redisPushQueue{})),
		client: asynq.NewClientFromRedisClient(redisClient),
		config: config,
	}
}

// Enqueue a task to the queue
func (queue *redisPushQueue) Enqueue(ctx context.Context, task *PushQueueTask, timeout time.Duration) (queueID string, err error) {
	_, span, ctxLogger := queue.tracer.StartWithLogger(ctx, queue.logger)
	defer span.End()

	payload, err := json.Marshal(task)
	if err != nil {
		return "", stacktrace.Propagate(err, fmt.Sprintf("cannot marshal push queue task for URL [%s]", task.URL))
	}

	info, err := queue.client.Enqueue(
		asynq.NewTask(RedisPushQueueTaskType, payload),
		asynq.ProcessIn(timeout),
		asynq.MaxRetry(3),
		asynq.Queue(queue.config.Name),
	)
	if err != nil {
		return "", stacktrace.Propagate(err, fmt.Sprintf("cannot enqueue task for URL [%s]", task.URL))
	}

	ctxLogger.Info(fmt.Sprintf(
		"task added to [%s] queue with ID [%s] and scheduled at [%s]",
		queue.config.Name,
		info.ID,
		time.Now().UTC().Add(timeout),
	))

	return info.ID, nil
}

// RedisPushQueueHandler executes the HTTP callback for tasks enqueued by RedisPushQueue.
// It runs as an in-process asynq worker; returning an error causes asynq to retry the task.
type RedisPushQueueHandler struct {
	client *http.Client
	logger telemetry.Logger
}

// NewRedisPushQueueHandler creates a new RedisPushQueueHandler
func NewRedisPushQueueHandler(logger telemetry.Logger, client *http.Client) *RedisPushQueueHandler {
	return &RedisPushQueueHandler{
		client: client,
		logger: logger.WithService(fmt.Sprintf("%T", RedisPushQueueHandler{})),
	}
}

// ProcessTask implements asynq.Handler
func (h *RedisPushQueueHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var task PushQueueTask
	if err := json.Unmarshal(t.Payload(), &task); err != nil {
		return stacktrace.Propagate(err, "cannot unmarshal push queue task payload")
	}

	requestCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	request := requests.
		URL(task.URL).
		Client(h.client).
		Method(task.Method).
		BodyBytes(task.Body).
		Header("Content-Type", "application/json")

	for key, value := range task.Headers {
		request.Header(key, value)
	}

	if err := request.Fetch(requestCtx); err != nil {
		return stacktrace.Propagate(err, fmt.Sprintf("cannot send http request to [%s]", task.URL))
	}

	h.logger.Info(fmt.Sprintf("push queue task sent to URL [%s]", task.URL))
	return nil
}
