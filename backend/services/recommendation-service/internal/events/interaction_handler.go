package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/usecase"
)

const (
	UnknownEventPolicyDLQ  = "dlq"
	UnknownEventPolicySkip = "skip"
)

type Message struct {
	Topic      string
	Key        []byte
	Value      []byte
	ReceivedAt time.Time
}

type InteractionIngestor interface {
	IngestInteractions(ctx context.Context, interactions []domain.UserInteraction) (usecase.InteractionIngestionResult, error)
}

type ConsumerMetrics interface {
	RecordConsumed(eventType string, producer string, result string)
	ObserveProcessingDuration(eventType string, duration time.Duration)
	RecordDLQ(reason string, eventType string)
}

type InteractionHandlerConfig struct {
	ServiceName        string
	Topic              string
	SupportedVersion   int
	MaxMessageBytes    int64
	MaxRetryAttempts   int
	RetryBackoffs      []time.Duration
	UnknownEventPolicy string
}

type InteractionHandler struct {
	mapper             *InteractionMapper
	ingestor           InteractionIngestor
	dlq                DeadLetterPublisher
	metrics            ConsumerMetrics
	logger             *slog.Logger
	serviceName        string
	topic              string
	supportedVersion   int
	maxMessageBytes    int64
	maxRetryAttempts   int
	retryBackoffs      []time.Duration
	unknownEventPolicy string
}

func NewInteractionHandler(
	mapper *InteractionMapper,
	ingestor InteractionIngestor,
	dlq DeadLetterPublisher,
	metrics ConsumerMetrics,
	cfg InteractionHandlerConfig,
	logger *slog.Logger,
) (*InteractionHandler, error) {
	if mapper == nil {
		return nil, errors.New("interaction mapper is required")
	}
	if ingestor == nil {
		return nil, errors.New("interaction ingestor is required")
	}
	if strings.TrimSpace(cfg.ServiceName) == "" {
		cfg.ServiceName = "recommendation-service"
	}
	if strings.TrimSpace(cfg.Topic) == "" {
		cfg.Topic = "recommendation.events"
	}
	if cfg.SupportedVersion <= 0 {
		cfg.SupportedVersion = DefaultSupportedEventVersion
	}
	if cfg.MaxMessageBytes <= 0 {
		cfg.MaxMessageBytes = 256 << 10
	}
	if cfg.MaxRetryAttempts <= 0 {
		cfg.MaxRetryAttempts = 1
	}
	if cfg.UnknownEventPolicy == "" {
		cfg.UnknownEventPolicy = UnknownEventPolicyDLQ
	}
	if cfg.UnknownEventPolicy != UnknownEventPolicyDLQ && cfg.UnknownEventPolicy != UnknownEventPolicySkip {
		return nil, fmt.Errorf("unsupported unknown event policy %q", cfg.UnknownEventPolicy)
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &InteractionHandler{
		mapper:             mapper,
		ingestor:           ingestor,
		dlq:                dlq,
		metrics:            metrics,
		logger:             logger,
		serviceName:        strings.TrimSpace(cfg.ServiceName),
		topic:              strings.TrimSpace(cfg.Topic),
		supportedVersion:   cfg.SupportedVersion,
		maxMessageBytes:    cfg.MaxMessageBytes,
		maxRetryAttempts:   cfg.MaxRetryAttempts,
		retryBackoffs:      append([]time.Duration(nil), cfg.RetryBackoffs...),
		unknownEventPolicy: cfg.UnknownEventPolicy,
	}, nil
}

func (h *InteractionHandler) Handle(ctx context.Context, message Message) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	start := time.Now()
	receivedAt := message.ReceivedAt
	if receivedAt.IsZero() {
		receivedAt = start.UTC()
	}

	eventTypeForMetrics := "unknown"
	producerForMetrics := "unknown"
	defer func() {
		if h.metrics != nil {
			h.metrics.ObserveProcessingDuration(eventTypeForMetrics, time.Since(start))
		}
	}()

	if int64(len(message.Value)) > h.maxMessageBytes {
		eventID, eventType := summarizeRawEvent(message.Value)
		if err := h.publishDLQ(ctx, dlqInput{
			raw:           message.Value,
			originalTopic: h.originalTopic(message.Topic),
			eventID:       eventID,
			eventType:     eventType,
			reason:        "message_too_large",
			attempts:      0,
		}); err != nil {
			return err
		}
		h.recordConsumed(eventTypeOrUnknown(eventType), producerForMetrics, "dlq")
		return nil
	}

	envelope, err := DecodeEnvelope(message.Value)
	if err != nil {
		eventID, eventType := summarizeRawEvent(message.Value)
		if eventType != "" {
			eventTypeForMetrics = eventType
		}
		if publishErr := h.publishDLQ(ctx, dlqInput{
			raw:           message.Value,
			originalTopic: h.originalTopic(message.Topic),
			eventID:       eventID,
			eventType:     eventType,
			reason:        err.Error(),
			attempts:      0,
		}); publishErr != nil {
			return publishErr
		}
		h.recordConsumed(eventTypeForMetrics, producerForMetrics, "dlq")
		return nil
	}
	eventTypeForMetrics = string(envelope.EventType)
	producerForMetrics = envelope.Producer

	if err := envelope.Validate(h.supportedVersion); err != nil {
		if errors.Is(err, domain.ErrUnsupportedEventType) && h.unknownEventPolicy == UnknownEventPolicySkip {
			h.logger.InfoContext(ctx, "recommendation.interaction.event_skipped",
				slog.String("event_id", envelope.EventID),
				slog.String("event_type", string(envelope.EventType)),
				slog.String("reason", "unsupported_event_type"),
			)
			h.recordConsumed(eventTypeForMetrics, producerForMetrics, "skipped")
			return nil
		}
		if publishErr := h.publishDLQ(ctx, dlqInput{
			raw:           message.Value,
			originalTopic: h.originalTopic(message.Topic),
			envelope:      envelope,
			reason:        err.Error(),
			attempts:      0,
		}); publishErr != nil {
			return publishErr
		}
		h.recordConsumed(eventTypeForMetrics, producerForMetrics, "dlq")
		return nil
	}

	interactions, err := h.mapper.Map(envelope, receivedAt)
	if err != nil {
		if publishErr := h.publishDLQ(ctx, dlqInput{
			raw:           message.Value,
			originalTopic: h.originalTopic(message.Topic),
			envelope:      envelope,
			reason:        err.Error(),
			attempts:      0,
		}); publishErr != nil {
			return publishErr
		}
		h.recordConsumed(eventTypeForMetrics, producerForMetrics, "dlq")
		return nil
	}

	result, attempts, err := h.ingestWithRetry(ctx, interactions)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		if publishErr := h.publishDLQ(ctx, dlqInput{
			raw:           message.Value,
			originalTopic: h.originalTopic(message.Topic),
			envelope:      envelope,
			reason:        err.Error(),
			attempts:      attempts,
		}); publishErr != nil {
			return publishErr
		}
		h.recordConsumed(eventTypeForMetrics, producerForMetrics, "dlq")
		return nil
	}

	outcome := "stored"
	if result.Stored == 0 && result.FeaturesApplied > 0 {
		outcome = "features_repaired"
	} else if result.Stored == 0 && result.Duplicates > 0 {
		outcome = "duplicate"
	}
	h.logger.InfoContext(ctx, "recommendation.interaction.event_processed",
		slog.String("event_id", envelope.EventID),
		slog.String("event_type", string(envelope.EventType)),
		slog.String("producer", envelope.Producer),
		slog.String("trace_id", envelope.TraceID),
		slog.Int("stored", result.Stored),
		slog.Int("duplicates", result.Duplicates),
		slog.Int("features_applied", result.FeaturesApplied),
		slog.Int("feature_duplicates", result.FeatureDuplicates),
		slog.Int("attempts", attempts),
		slog.String("result", outcome),
	)
	h.recordConsumed(eventTypeForMetrics, producerForMetrics, outcome)
	return nil
}

func (h *InteractionHandler) ingestWithRetry(ctx context.Context, interactions []domain.UserInteraction) (usecase.InteractionIngestionResult, int, error) {
	var lastErr error
	for attempt := 1; attempt <= h.maxRetryAttempts; attempt++ {
		result, err := h.ingestor.IngestInteractions(ctx, interactions)
		if err == nil {
			return result, attempt, nil
		}
		if !retryableIngestionError(err) {
			return result, attempt, err
		}
		lastErr = err
		if attempt == h.maxRetryAttempts {
			break
		}
		if err := sleepBackoff(ctx, h.backoff(attempt)); err != nil {
			return result, attempt, err
		}
	}
	return usecase.InteractionIngestionResult{}, h.maxRetryAttempts, fmt.Errorf("%w: max retry attempts exceeded: %v", domain.ErrRecommendationStorage, lastErr)
}

func retryableIngestionError(err error) bool {
	return errors.Is(err, domain.ErrRecommendationStorage)
}

func sleepBackoff(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (h *InteractionHandler) backoff(attempt int) time.Duration {
	if len(h.retryBackoffs) == 0 {
		return 0
	}
	index := attempt - 1
	if index >= len(h.retryBackoffs) {
		index = len(h.retryBackoffs) - 1
	}
	return h.retryBackoffs[index]
}

type dlqInput struct {
	raw           []byte
	originalTopic string
	envelope      Envelope
	eventID       string
	eventType     string
	reason        string
	attempts      int
}

func (h *InteractionHandler) publishDLQ(ctx context.Context, input dlqInput) error {
	if h.dlq == nil {
		return errors.New("dead-letter publisher is required")
	}
	eventID := input.eventID
	eventType := input.eventType
	producer := ""
	traceID := ""
	if input.envelope.EventID != "" {
		eventID = input.envelope.EventID
		eventType = string(input.envelope.EventType)
		producer = input.envelope.Producer
		traceID = input.envelope.TraceID
	}
	message := DeadLetterMessage{
		FailedAt:      time.Now().UTC(),
		Reason:        input.reason,
		Consumer:      h.serviceName,
		OriginalTopic: input.originalTopic,
		EventID:       eventID,
		EventType:     eventType,
		Producer:      producer,
		TraceID:       traceID,
		Attempts:      input.attempts,
	}
	if json.Valid(input.raw) {
		message.OriginalEvent = append([]byte(nil), input.raw...)
	} else {
		message.OriginalEventRaw = string(input.raw)
	}
	if err := h.dlq.Publish(ctx, message); err != nil {
		return fmt.Errorf("publish dead-letter event: %w", err)
	}
	if h.metrics != nil {
		h.metrics.RecordDLQ(dlqReasonLabel(input.reason), eventTypeOrUnknown(eventType))
	}
	h.logger.WarnContext(ctx, "recommendation.interaction.event_dead_lettered",
		slog.String("event_id", eventID),
		slog.String("event_type", eventTypeOrUnknown(eventType)),
		slog.String("reason", input.reason),
		slog.Int("attempts", input.attempts),
	)
	return nil
}

func (h *InteractionHandler) recordConsumed(eventType string, producer string, result string) {
	if h.metrics != nil {
		h.metrics.RecordConsumed(eventTypeOrUnknown(eventType), producerOrUnknown(producer), result)
	}
}

func (h *InteractionHandler) originalTopic(topic string) string {
	topic = strings.TrimSpace(topic)
	if topic != "" {
		return topic
	}
	return h.topic
}

func summarizeRawEvent(raw []byte) (string, string) {
	var partial struct {
		EventID   string `json:"event_id"`
		EventType string `json:"event_type"`
	}
	if err := json.Unmarshal(raw, &partial); err != nil {
		return "", ""
	}
	return strings.TrimSpace(partial.EventID), strings.TrimSpace(partial.EventType)
}

func eventTypeOrUnknown(eventType string) string {
	eventType = strings.TrimSpace(eventType)
	if eventType == "" {
		return "unknown"
	}
	return eventType
}

func producerOrUnknown(producer string) string {
	producer = strings.TrimSpace(producer)
	if producer == "" {
		return "unknown"
	}
	return producer
}

func dlqReasonLabel(reason string) string {
	reason = strings.ToLower(strings.TrimSpace(reason))
	switch {
	case reason == "":
		return "unknown"
	case strings.Contains(reason, "message_too_large"):
		return "message_too_large"
	case strings.Contains(reason, "invalid json"):
		return "invalid_json"
	case strings.Contains(reason, "unsupported event version"):
		return "unsupported_version"
	case strings.Contains(reason, "unsupported event"):
		return "unsupported_event"
	case strings.Contains(reason, "invalid payload"):
		return "invalid_payload"
	case strings.Contains(reason, "max retry attempts"):
		return "retry_exhausted"
	case strings.Contains(reason, "product_id"):
		return "missing_or_invalid_product_id"
	case strings.Contains(reason, "user_id or anonymous_id"):
		return "missing_identity"
	default:
		return "validation_error"
	}
}
