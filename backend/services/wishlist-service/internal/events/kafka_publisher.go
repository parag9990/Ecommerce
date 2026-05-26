package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"ecommerce/backend/services/wishlist-service/internal/domain"

	"github.com/segmentio/kafka-go"
)

const WishlistAnalyticsProducer = "wishlist-service"

type KafkaWishlistAnalyticsPublisherConfig struct {
	Brokers []string
	Topic   string
}

type WishlistAnalyticsEventEnvelope struct {
	EventID    string                        `json:"event_id"`
	EventType  domain.WishlistEventType      `json:"event_type"`
	Version    int                           `json:"version"`
	OccurredAt time.Time                     `json:"occurred_at"`
	Producer   string                        `json:"producer"`
	TraceID    string                        `json:"trace_id,omitempty"`
	Payload    WishlistAnalyticsEventPayload `json:"payload"`
}

type WishlistAnalyticsEventPayload struct {
	UserID         string                     `json:"user_id"`
	ProductID      string                     `json:"product_id"`
	VariantID      string                     `json:"variant_id,omitempty"`
	Action         domain.WishlistEventAction `json:"action"`
	Source         string                     `json:"source"`
	Availability   domain.Availability        `json:"availability,omitempty"`
	LastKnownPrice *WishlistAnalyticsMoney    `json:"last_known_price,omitempty"`
}

type WishlistAnalyticsMoney struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

type kafkaWishlistAnalyticsWriter interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

type KafkaWishlistAnalyticsPublisher struct {
	config KafkaWishlistAnalyticsPublisherConfig
	writer kafkaWishlistAnalyticsWriter
	logger *slog.Logger
}

func NewKafkaWishlistAnalyticsPublisher(config KafkaWishlistAnalyticsPublisherConfig, logger *slog.Logger) (*KafkaWishlistAnalyticsPublisher, error) {
	normalized, err := normalizeKafkaWishlistAnalyticsPublisherConfig(config)
	if err != nil {
		return nil, err
	}
	if logger == nil {
		logger = slog.Default()
	}
	writer := &kafka.Writer{
		Addr:         kafka.TCP(normalized.Brokers...),
		Topic:        normalized.Topic,
		Balancer:     &kafka.Hash{},
		RequiredAcks: kafka.RequireAll,
	}
	return newKafkaWishlistAnalyticsPublisherWithWriter(normalized, writer, logger), nil
}

func newKafkaWishlistAnalyticsPublisherWithWriter(config KafkaWishlistAnalyticsPublisherConfig, writer kafkaWishlistAnalyticsWriter, logger *slog.Logger) *KafkaWishlistAnalyticsPublisher {
	if logger == nil {
		logger = slog.Default()
	}
	return &KafkaWishlistAnalyticsPublisher{
		config: config,
		writer: writer,
		logger: logger,
	}
}

func (p *KafkaWishlistAnalyticsPublisher) Publish(ctx context.Context, event domain.WishlistAnalyticsEvent) error {
	if p == nil || p.writer == nil {
		return errors.New("wishlist analytics publisher is not initialized")
	}
	event = event.Normalized()
	if strings.TrimSpace(event.Topic) == "" {
		event.Topic = p.config.Topic
	}
	if err := event.Validate(); err != nil {
		return err
	}

	envelope := WishlistAnalyticsEventEnvelope{
		EventID:    event.EventID,
		EventType:  event.EventType,
		Version:    event.Version,
		OccurredAt: event.OccurredAt,
		Producer:   WishlistAnalyticsProducer,
		TraceID:    event.TraceID,
		Payload: WishlistAnalyticsEventPayload{
			UserID:         event.Payload.UserID,
			ProductID:      event.Payload.ProductID,
			VariantID:      event.Payload.VariantID,
			Action:         event.Payload.Action,
			Source:         event.Payload.Source,
			Availability:   event.Payload.Availability,
			LastKnownPrice: wishlistAnalyticsMoney(event.Payload.LastKnownPrice),
		},
	}
	body, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("encode wishlist analytics event: %w", err)
	}

	if err := p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(event.Payload.UserID),
		Value: body,
		Time:  event.OccurredAt,
		Headers: []kafka.Header{
			{Key: "event_id", Value: []byte(event.EventID)},
			{Key: "event_type", Value: []byte(event.EventType)},
			{Key: "producer", Value: []byte(WishlistAnalyticsProducer)},
		},
	}); err != nil {
		return fmt.Errorf("publish wishlist analytics event %q to %q: %w", event.EventID, p.config.Topic, err)
	}

	p.logger.Debug("wishlist analytics event published",
		"event_id", event.EventID,
		"event_type", event.EventType,
		"user_id", event.Payload.UserID,
		"product_id", event.Payload.ProductID,
		"topic", p.config.Topic,
	)
	return nil
}

func (p *KafkaWishlistAnalyticsPublisher) Close(context.Context) error {
	if p == nil || p.writer == nil {
		return nil
	}
	return p.writer.Close()
}

func normalizeKafkaWishlistAnalyticsPublisherConfig(config KafkaWishlistAnalyticsPublisherConfig) (KafkaWishlistAnalyticsPublisherConfig, error) {
	normalized := config
	normalized.Brokers = normalizeStringSlice(normalized.Brokers)
	normalized.Topic = strings.TrimSpace(normalized.Topic)

	var validationErrors []string
	if len(normalized.Brokers) == 0 {
		validationErrors = append(validationErrors, "kafka brokers are required")
	}
	if normalized.Topic == "" {
		validationErrors = append(validationErrors, "wishlist analytics topic is required")
	}
	if len(validationErrors) > 0 {
		return KafkaWishlistAnalyticsPublisherConfig{}, errors.New(strings.Join(validationErrors, "; "))
	}
	return normalized, nil
}

func wishlistAnalyticsMoney(money *domain.Money) *WishlistAnalyticsMoney {
	if money == nil {
		return nil
	}
	return &WishlistAnalyticsMoney{
		Amount:   money.Amount,
		Currency: money.Currency,
	}
}
