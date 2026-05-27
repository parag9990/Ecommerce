package repository

import (
	"context"
	"time"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
)

type TemplateRepository interface {
	FindLatestActive(ctx context.Context, templateKey string, channel domain.Channel) (domain.Template, error)
	InsertTemplate(ctx context.Context, template domain.Template) error
}

type DeliveryRepository interface {
	FindDeliveryByID(ctx context.Context, id string) (domain.Delivery, error)
	InsertDelivery(ctx context.Context, delivery domain.Delivery) error
}

type PreferenceRepository interface {
	FindPreferenceByUserID(ctx context.Context, userID string) (domain.Preference, error)
	UpsertPreference(ctx context.Context, preference domain.Preference) (domain.Preference, error)
}

type EventDeliveryRepository interface {
	ClaimEventDelivery(ctx context.Context, delivery domain.Delivery) (bool, error)
	RecordEventDeliveryOutcome(ctx context.Context, idempotencyKey string, outcome domain.EventDeliveryOutcome) error
}

type RetryDeliveryRepository interface {
	PrepareEventDelivery(ctx context.Context, delivery domain.Delivery) (bool, error)
	BeginAttempt(ctx context.Context, deliveryID string, idempotencyKey string, attempt int, leaseUntil time.Time, updatedAt time.Time) (domain.AttemptClaim, error)
	MarkAccepted(ctx context.Context, deliveryID string, attempt int, providerName, providerMessageID string, updatedAt time.Time) error
	MarkSuppressed(ctx context.Context, deliveryID string, attempt int, reason domain.ConsentReason, updatedAt time.Time) error
	MarkRetryScheduled(ctx context.Context, deliveryID string, attempt int, nextRetryAt time.Time, failureCode string, updatedAt time.Time) error
	MarkDeadLettered(ctx context.Context, deliveryID string, attempt int, failureCode string, deadLetteredAt time.Time) error
}

type AnalyticsRepository interface {
	FindDeliveryByProviderMessageID(ctx context.Context, provider, messageID string, channel domain.Channel) (domain.Delivery, error)
	RecordEventAndApplyMilestone(ctx context.Context, event domain.ProviderEvent) (domain.DeliveryEventApplyResult, error)
}
