package analytics

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
)

type Store interface {
	FindDeliveryByID(ctx context.Context, id string) (domain.Delivery, error)
	FindDeliveryByProviderMessageID(ctx context.Context, provider, messageID string, channel domain.Channel) (domain.Delivery, error)
	RecordEventAndApplyMilestone(ctx context.Context, event domain.ProviderEvent) (domain.DeliveryEventApplyResult, error)
}

type Observer interface {
	ObserveDeliveryEvent(event domain.ProviderEvent)
	ObserveWebhook(provider, outcome string)
	ObserveWebhookDuration(provider string, duration time.Duration)
}

type Recorder struct {
	store   Store
	metrics Observer
	now     func() time.Time
}

func NewRecorder(store Store, metrics Observer) (*Recorder, error) {
	if store == nil || metrics == nil {
		return nil, errors.New("delivery analytics dependencies are required")
	}
	return &Recorder{store: store, metrics: metrics, now: time.Now}, nil
}

func (r *Recorder) WithClock(now func() time.Time) {
	if now != nil {
		r.now = now
	}
}

func (r *Recorder) RecordSent(
	ctx context.Context,
	deliveryID string,
	provider string,
	providerMessageID string,
	occurredAt time.Time,
) (domain.DeliveryEventApplyResult, error) {
	if strings.TrimSpace(deliveryID) == "" || strings.TrimSpace(provider) == "" ||
		strings.TrimSpace(providerMessageID) == "" || occurredAt.IsZero() {
		return domain.DeliveryEventApplyResult{}, fmt.Errorf("%w: accepted delivery identity is required", domain.ErrInvalidProviderEvent)
	}
	delivery, err := r.store.FindDeliveryByID(ctx, strings.TrimSpace(deliveryID))
	if err != nil {
		return domain.DeliveryEventApplyResult{}, err
	}
	return r.record(ctx, eventFromDelivery(delivery, strings.TrimSpace(provider), "sent:"+delivery.ID,
		strings.TrimSpace(providerMessageID), domain.DeliveryEventSent, occurredAt, r.now().UTC(), ""))
}

func (r *Recorder) RecordFailed(
	ctx context.Context,
	deliveryID string,
	providerName string,
	failureCode string,
	occurredAt time.Time,
) (domain.DeliveryEventApplyResult, error) {
	if strings.TrimSpace(deliveryID) == "" || occurredAt.IsZero() {
		return domain.DeliveryEventApplyResult{}, fmt.Errorf("%w: failed delivery identity is required", domain.ErrInvalidProviderEvent)
	}
	delivery, err := r.store.FindDeliveryByID(ctx, strings.TrimSpace(deliveryID))
	if err != nil {
		return domain.DeliveryEventApplyResult{}, err
	}
	provider := strings.TrimSpace(providerName)
	if provider == "" {
		provider = strings.TrimSpace(delivery.Provider)
	}
	if provider == "" {
		provider = "notification_retry"
	}
	code := safeFailureCode(failureCode)
	return r.record(ctx, eventFromDelivery(delivery, provider, "failed:"+delivery.ID,
		delivery.ProviderMessageID, domain.DeliveryEventFailed, occurredAt, r.now().UTC(), code))
}

func (r *Recorder) RecordCallback(
	ctx context.Context,
	callback domain.ProviderCallback,
) (domain.DeliveryEventApplyResult, error) {
	if err := callback.Validate(); err != nil {
		return domain.DeliveryEventApplyResult{}, err
	}
	delivery, err := r.store.FindDeliveryByProviderMessageID(
		ctx,
		strings.TrimSpace(callback.Provider),
		strings.TrimSpace(callback.ProviderMessageID),
		callback.Channel,
	)
	if err != nil {
		return domain.DeliveryEventApplyResult{}, err
	}
	if callback.Type == domain.DeliveryEventOpened &&
		delivery.TemplateKey == string(domain.TemplateOTPVerification) {
		return domain.DeliveryEventApplyResult{}, fmt.Errorf("%w: observed opens are not recorded for OTP deliveries", domain.ErrInvalidProviderEvent)
	}
	failureCode := ""
	if callback.Type == domain.DeliveryEventFailed {
		failureCode = safeFailureCode(callback.FailureCode)
	}
	return r.record(ctx, eventFromDelivery(delivery, callback.Provider, callback.ProviderEventID,
		callback.ProviderMessageID, callback.Type, callback.OccurredAt, r.now().UTC(), failureCode))
}

func (r *Recorder) record(
	ctx context.Context,
	event domain.ProviderEvent,
) (domain.DeliveryEventApplyResult, error) {
	if ctx == nil {
		return domain.DeliveryEventApplyResult{}, errors.New("delivery analytics context is required")
	}
	if err := ctx.Err(); err != nil {
		return domain.DeliveryEventApplyResult{}, err
	}
	if err := event.Validate(); err != nil {
		return domain.DeliveryEventApplyResult{}, err
	}
	result, err := r.store.RecordEventAndApplyMilestone(ctx, event)
	if err != nil {
		return domain.DeliveryEventApplyResult{}, err
	}
	if result.MilestoneChanged {
		r.metrics.ObserveDeliveryEvent(event)
	}
	return result, nil
}

func eventFromDelivery(
	delivery domain.Delivery,
	provider string,
	providerEventID string,
	providerMessageID string,
	eventType domain.DeliveryEventType,
	occurredAt time.Time,
	receivedAt time.Time,
	failureCode string,
) domain.ProviderEvent {
	return domain.ProviderEvent{
		ID:                providerEventDocumentID(provider, providerEventID),
		Provider:          strings.TrimSpace(provider),
		ProviderEventID:   strings.TrimSpace(providerEventID),
		ProviderMessageID: strings.TrimSpace(providerMessageID),
		DeliveryID:        delivery.ID,
		Type:              eventType,
		Channel:           delivery.Channel,
		TemplateKey:       delivery.TemplateKey,
		CampaignID:        delivery.CampaignID,
		OccurredAt:        occurredAt.UTC(),
		ReceivedAt:        receivedAt.UTC(),
		FailureCode:       failureCode,
	}
}

func providerEventDocumentID(provider, eventID string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(provider) + ":" + strings.TrimSpace(eventID)))
	return "provider_event_" + hex.EncodeToString(sum[:16])
}

var failureCodeCharacters = regexp.MustCompile(`[^a-z0-9_:-]+`)

func safeFailureCode(code string) string {
	code = strings.ToLower(strings.TrimSpace(code))
	code = failureCodeCharacters.ReplaceAllString(code, "_")
	code = strings.Trim(code, "_:")
	if code == "" {
		return "provider_terminal_failure"
	}
	if len(code) > 64 {
		return code[:64]
	}
	return code
}
