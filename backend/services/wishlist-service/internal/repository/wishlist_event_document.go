package repository

import (
	"fmt"
	"strings"
	"time"

	"ecommerce/backend/services/wishlist-service/internal/domain"
)

type WishlistAnalyticsPayloadDocument struct {
	UserID         string                     `bson:"user_id" json:"user_id"`
	ProductID      string                     `bson:"product_id" json:"product_id"`
	VariantID      string                     `bson:"variant_id,omitempty" json:"variant_id,omitempty"`
	Action         domain.WishlistEventAction `bson:"action" json:"action"`
	Source         string                     `bson:"source" json:"source"`
	Availability   domain.Availability        `bson:"availability,omitempty" json:"availability,omitempty"`
	LastKnownPrice *MoneySnapshotDocument     `bson:"last_known_price,omitempty" json:"last_known_price,omitempty"`
}

type WishlistEventDocument struct {
	ID          string                           `bson:"_id" json:"event_id"`
	EventType   domain.WishlistEventType         `bson:"event_type" json:"event_type"`
	Version     int32                            `bson:"version" json:"version"`
	Topic       string                           `bson:"topic" json:"topic"`
	Payload     WishlistAnalyticsPayloadDocument `bson:"payload" json:"payload"`
	TraceID     string                           `bson:"trace_id,omitempty" json:"trace_id,omitempty"`
	Status      domain.WishlistEventStatus       `bson:"status" json:"status"`
	Attempts    int32                            `bson:"attempts" json:"attempts"`
	NextRetryAt time.Time                        `bson:"next_retry_at" json:"next_retry_at"`
	LastError   string                           `bson:"last_error,omitempty" json:"last_error,omitempty"`
	OccurredAt  time.Time                        `bson:"occurred_at" json:"occurred_at"`
	PublishedAt *time.Time                       `bson:"published_at,omitempty" json:"published_at,omitempty"`
	CreatedAt   time.Time                        `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time                        `bson:"updated_at" json:"updated_at"`
}

func NewWishlistEventDocument(event domain.WishlistAnalyticsEvent) (WishlistEventDocument, error) {
	event = event.Normalized()
	if err := event.Validate(); err != nil {
		return WishlistEventDocument{}, err
	}
	return WishlistEventDocument{
		ID:        event.EventID,
		EventType: event.EventType,
		Version:   int32(event.Version),
		Topic:     event.Topic,
		Payload: WishlistAnalyticsPayloadDocument{
			UserID:         event.Payload.UserID,
			ProductID:      event.Payload.ProductID,
			VariantID:      event.Payload.VariantID,
			Action:         event.Payload.Action,
			Source:         event.Payload.Source,
			Availability:   event.Payload.Availability,
			LastKnownPrice: newMoneySnapshotDocument(event.Payload.LastKnownPrice),
		},
		TraceID:     event.TraceID,
		Status:      event.Status,
		Attempts:    int32(event.Attempts),
		NextRetryAt: event.NextRetryAt,
		LastError:   truncateWishlistEventError(event.LastError),
		OccurredAt:  event.OccurredAt,
		PublishedAt: copyTime(event.PublishedAt),
		CreatedAt:   event.CreatedAt,
		UpdatedAt:   event.UpdatedAt,
	}, nil
}

func (d WishlistEventDocument) ToDomain() (domain.WishlistAnalyticsEvent, error) {
	event := domain.WishlistAnalyticsEvent{
		EventID:   d.ID,
		EventType: d.EventType,
		Version:   int(d.Version),
		Topic:     d.Topic,
		Payload: domain.WishlistAnalyticsPayload{
			UserID:         d.Payload.UserID,
			ProductID:      d.Payload.ProductID,
			VariantID:      d.Payload.VariantID,
			Action:         d.Payload.Action,
			Source:         d.Payload.Source,
			Availability:   d.Payload.Availability,
			LastKnownPrice: d.Payload.LastKnownPrice.toDomain(),
		},
		TraceID:     d.TraceID,
		Status:      d.Status,
		Attempts:    int(d.Attempts),
		NextRetryAt: d.NextRetryAt,
		LastError:   d.LastError,
		OccurredAt:  d.OccurredAt,
		PublishedAt: copyTime(d.PublishedAt),
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}.Normalized()
	if err := event.Validate(); err != nil {
		return domain.WishlistAnalyticsEvent{}, fmt.Errorf("decode wishlist analytics event %q: %w", d.ID, err)
	}
	return event, nil
}

func copyTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	copied := value.UTC()
	return &copied
}

func truncateWishlistEventError(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= maxWishlistEventLastErrorLength {
		return value
	}
	return value[:maxWishlistEventLastErrorLength]
}
