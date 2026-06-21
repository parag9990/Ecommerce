package domain

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

// DeliveryStatus records provider-attempt state without implying user receipt.
type DeliveryStatus string

const (
	DeliveryStatusProcessing     DeliveryStatus = "processing"
	DeliveryStatusPending        DeliveryStatus = "pending"
	DeliveryStatusAccepted       DeliveryStatus = "accepted"
	DeliveryStatusRejected       DeliveryStatus = "rejected"
	DeliveryStatusFailed         DeliveryStatus = "failed"
	DeliveryStatusSent           DeliveryStatus = "sent"
	DeliveryStatusRetryScheduled DeliveryStatus = "retry_scheduled"
	DeliveryStatusDeadLettered   DeliveryStatus = "dead_lettered"
	DeliveryStatusSuppressed     DeliveryStatus = "suppressed"
)

type Delivery struct {
	ID                   string         `bson:"_id" json:"id"`
	UserID               string         `bson:"user_id,omitempty" json:"user_id,omitempty"`
	Channel              Channel        `bson:"channel" json:"channel"`
	TemplateKey          string         `bson:"template_key" json:"template_key"`
	CampaignID           string         `bson:"campaign_id,omitempty" json:"campaign_id,omitempty"`
	Status               DeliveryStatus `bson:"status" json:"status"`
	Provider             string         `bson:"provider,omitempty" json:"provider,omitempty"`
	ProviderMessageID    string         `bson:"provider_message_id,omitempty" json:"provider_message_id,omitempty"`
	Attempts             int            `bson:"attempts" json:"attempts"`
	Payload              map[string]any `bson:"payload,omitempty" json:"payload,omitempty"`
	IdempotencyKey       string         `bson:"idempotency_key,omitempty" json:"idempotency_key,omitempty"`
	SourceEventID        string         `bson:"source_event_id,omitempty" json:"source_event_id,omitempty"`
	SourceEventType      string         `bson:"source_event_type,omitempty" json:"source_event_type,omitempty"`
	TraceID              string         `bson:"trace_id,omitempty" json:"trace_id,omitempty"`
	RecipientCiphertext  string         `bson:"recipient_ciphertext,omitempty" json:"-"`
	MaxAttempts          int            `bson:"max_attempts,omitempty" json:"max_attempts,omitempty"`
	LastFailureCode      string         `bson:"last_failure_code,omitempty" json:"last_failure_code,omitempty"`
	SuppressionReason    ConsentReason  `bson:"suppression_reason,omitempty" json:"suppression_reason,omitempty"`
	NextRetryAt          *time.Time     `bson:"next_retry_at,omitempty" json:"next_retry_at,omitempty"`
	DeadLetteredAt       *time.Time     `bson:"dead_lettered_at,omitempty" json:"dead_lettered_at,omitempty"`
	SentAt               *time.Time     `bson:"sent_at,omitempty" json:"sent_at,omitempty"`
	DeliveredAt          *time.Time     `bson:"delivered_at,omitempty" json:"delivered_at,omitempty"`
	FailedAt             *time.Time     `bson:"failed_at,omitempty" json:"failed_at,omitempty"`
	OpenedAt             *time.Time     `bson:"opened_at,omitempty" json:"opened_at,omitempty"`
	FailureCode          string         `bson:"failure_code,omitempty" json:"failure_code,omitempty"`
	ProcessingAttempt    int            `bson:"processing_attempt,omitempty" json:"processing_attempt,omitempty"`
	ProcessingLeaseUntil *time.Time     `bson:"processing_lease_until,omitempty" json:"processing_lease_until,omitempty"`
	CreatedAt            time.Time      `bson:"created_at" json:"created_at"`
	UpdatedAt            time.Time      `bson:"updated_at" json:"updated_at"`
}

func (d Delivery) Validate() error {
	for value, field := range map[string]string{
		d.ID:             "id",
		d.TemplateKey:    "template_key",
		string(d.Status): "status",
	} {
		if err := requiredText(value, field); err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidDelivery, err)
		}
	}
	if blank(d.UserID) && d.TemplateKey != string(TemplateOTPVerification) {
		return fmt.Errorf("%w: user_id is required except for OTP deliveries", ErrInvalidDelivery)
	}
	if !d.Channel.IsSupported() {
		return fmt.Errorf("%w: %w: %q", ErrInvalidDelivery, ErrUnsupportedChannel, d.Channel)
	}
	if !d.Status.IsSupported() {
		return fmt.Errorf("%w: unsupported status %q", ErrInvalidDelivery, d.Status)
	}
	if d.Attempts < 0 {
		return fmt.Errorf("%w: attempts cannot be negative", ErrInvalidDelivery)
	}
	if d.MaxAttempts < 0 || (d.MaxAttempts > 0 && d.Attempts > d.MaxAttempts) {
		return fmt.Errorf("%w: invalid maximum attempt bounds", ErrInvalidDelivery)
	}
	if !blank(d.ProviderMessageID) && blank(d.Provider) {
		return fmt.Errorf("%w: provider is required with provider message id", ErrInvalidDelivery)
	}
	if err := validateRecordTimestamps(d.CreatedAt, d.UpdatedAt); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidDelivery, err)
	}
	if err := validatePayloadValue(reflect.ValueOf(d.Payload)); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidDelivery, err)
	}
	if err := d.validateRetryLifecycle(); err != nil {
		return err
	}
	if err := d.validateAnalyticsLifecycle(); err != nil {
		return err
	}
	if d.Status == DeliveryStatusSuppressed && !d.SuppressionReason.IsSuppressionReason() {
		return fmt.Errorf("%w: suppressed delivery requires suppression_reason", ErrInvalidDelivery)
	}
	if d.Status != DeliveryStatusSuppressed && strings.TrimSpace(string(d.SuppressionReason)) != "" {
		return fmt.Errorf("%w: suppression_reason requires suppressed status", ErrInvalidDelivery)
	}
	if err := d.validateSourceEvent(); err != nil {
		return err
	}
	return nil
}

func (d Delivery) validateAnalyticsLifecycle() error {
	for field, timestamp := range map[string]*time.Time{
		"sent_at":      d.SentAt,
		"delivered_at": d.DeliveredAt,
		"failed_at":    d.FailedAt,
		"opened_at":    d.OpenedAt,
	} {
		if timestamp != nil && timestamp.IsZero() {
			return fmt.Errorf("%w: %s must not be zero", ErrInvalidDelivery, field)
		}
	}
	if strings.TrimSpace(d.FailureCode) != "" && d.FailedAt == nil {
		return fmt.Errorf("%w: failure_code requires failed_at", ErrInvalidDelivery)
	}
	if d.FailedAt != nil && strings.TrimSpace(d.FailureCode) == "" {
		return fmt.Errorf("%w: failed_at requires failure_code", ErrInvalidDelivery)
	}
	return nil
}

func (d Delivery) validateSourceEvent() error {
	hasEventField := !blank(d.IdempotencyKey) || !blank(d.SourceEventID) ||
		!blank(d.SourceEventType) || !blank(d.TraceID)
	if !hasEventField {
		return nil
	}
	for value, field := range map[string]string{
		d.IdempotencyKey:  "idempotency_key",
		d.SourceEventID:   "source_event_id",
		d.SourceEventType: "source_event_type",
	} {
		if blank(value) {
			return fmt.Errorf("%w: %s is required for event deliveries", ErrInvalidDelivery, field)
		}
	}
	if d.TemplateKey == string(TemplateOTPVerification) {
		return fmt.Errorf("%w: OTP deliveries cannot be source-event deliveries", ErrInvalidDelivery)
	}
	return nil
}

func (d Delivery) validateRetryLifecycle() error {
	hasRetryField := d.MaxAttempts > 0 || !blank(d.RecipientCiphertext) ||
		!blank(d.LastFailureCode) || d.NextRetryAt != nil || d.DeadLetteredAt != nil ||
		d.ProcessingAttempt > 0 || d.ProcessingLeaseUntil != nil
	if !hasRetryField {
		return nil
	}
	if d.TemplateKey == string(TemplateOTPVerification) {
		return fmt.Errorf("%w: %w", ErrInvalidDelivery, ErrDurableOTPRetryForbidden)
	}
	if blank(d.IdempotencyKey) || blank(d.RecipientCiphertext) || d.MaxAttempts < 1 {
		return fmt.Errorf("%w: retryable event delivery requires identity, encrypted recipient, and max attempts",
			ErrInvalidDelivery)
	}
	if d.Status == DeliveryStatusProcessing {
		if d.ProcessingAttempt < 1 || d.ProcessingLeaseUntil == nil {
			return fmt.Errorf("%w: processing retry delivery requires an active attempt lease", ErrInvalidDelivery)
		}
	}
	if d.Status == DeliveryStatusRetryScheduled && (d.NextRetryAt == nil || d.Attempts < 1) {
		return fmt.Errorf("%w: scheduled retry requires attempt count and next retry time", ErrInvalidDelivery)
	}
	if d.Status == DeliveryStatusDeadLettered && d.DeadLetteredAt == nil {
		return fmt.Errorf("%w: dead-lettered delivery requires a terminal timestamp", ErrInvalidDelivery)
	}
	return nil
}

func (s DeliveryStatus) IsSupported() bool {
	switch s {
	case DeliveryStatusProcessing, DeliveryStatusPending, DeliveryStatusAccepted,
		DeliveryStatusRejected, DeliveryStatusFailed, DeliveryStatusSent,
		DeliveryStatusRetryScheduled, DeliveryStatusDeadLettered, DeliveryStatusSuppressed:
		return true
	default:
		return false
	}
}

func validatePayloadValue(value reflect.Value) error {
	if !value.IsValid() {
		return nil
	}
	for value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil
		}
		value = value.Elem()
	}
	switch value.Kind() {
	case reflect.Map:
		if value.Type().Key().Kind() != reflect.String {
			return nil
		}
		iter := value.MapRange()
		for iter.Next() {
			key := iter.Key().String()
			if sensitivePayloadKey(key) {
				return fmt.Errorf("payload field %q may contain sensitive data", key)
			}
			if err := validatePayloadValue(iter.Value()); err != nil {
				return err
			}
		}
	case reflect.Array, reflect.Slice:
		for i := 0; i < value.Len(); i++ {
			if err := validatePayloadValue(value.Index(i)); err != nil {
				return err
			}
		}
	}
	return nil
}

func sensitivePayloadKey(key string) bool {
	key = strings.ToLower(strings.TrimSpace(key))
	key = strings.NewReplacer("-", "_", ".", "_", " ", "_").Replace(key)
	for _, prohibited := range []string{"otp", "password", "secret", "token", "authorization", "api_key", "apikey"} {
		if strings.Contains(key, prohibited) {
			return true
		}
	}
	return false
}
