package sessionlink

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/domain"
)

const (
	ProducerAuthService = "auth-service"
	EventVersion        = 1

	EventTypeAuthLoginSucceeded        = "AuthLoginSucceeded"
	EventTypeAuthSignupSucceeded       = "AuthSignupSucceeded"
	EventTypeAuthLogoutSucceeded       = "AuthLogoutSucceeded"
	EventTypeRefreshTokenReuseDetected = "RefreshTokenReuseDetected"

	RoutingKeyAuthLoginSucceeded        = "auth.login_succeeded"
	RoutingKeyAuthSignupSucceeded       = "auth.signup_succeeded"
	RoutingKeyAuthLogoutSucceeded       = "auth.logout_succeeded"
	RoutingKeyRefreshTokenReuseDetected = "auth.refresh_token_reuse_detected"

	AggregateTypeSession = "session"
	AggregateTypeAccount = "account"

	AuthMethodPassword = "password"
	AuthMethodSignup   = "signup"

	LogoutReasonUserRequested  = "user_requested"
	ActionRevokedSessionFamily = "revoked_session_family"
)

var ErrInvalidEvent = errors.New("invalid session link event")

type EventEnvelope struct {
	EventID       string          `json:"event_id"`
	EventType     string          `json:"event_type"`
	Version       int             `json:"version"`
	OccurredAt    time.Time       `json:"occurred_at"`
	Producer      string          `json:"producer"`
	TraceID       string          `json:"trace_id,omitempty"`
	AggregateType string          `json:"aggregate_type"`
	AggregateID   string          `json:"aggregate_id"`
	Payload       json.RawMessage `json:"payload"`
}

type LoginSucceededEvent struct {
	EventID     string      `json:"-"`
	TraceID     string      `json:"-"`
	OccurredAt  time.Time   `json:"-"`
	AccountID   string      `json:"account_id"`
	UserID      string      `json:"user_id"`
	SessionID   string      `json:"session_id"`
	AnonymousID string      `json:"anonymous_id,omitempty"`
	Roles       []string    `json:"roles,omitempty"`
	SellerID    string      `json:"seller_id,omitempty"`
	TenantID    string      `json:"tenant_id,omitempty"`
	Device      DeviceInfo  `json:"device"`
	Network     NetworkInfo `json:"network"`
	Auth        AuthInfo    `json:"auth"`
}

type SignupSucceededEvent struct {
	EventID     string      `json:"-"`
	TraceID     string      `json:"-"`
	OccurredAt  time.Time   `json:"-"`
	AccountID   string      `json:"account_id"`
	UserID      string      `json:"user_id"`
	SessionID   string      `json:"session_id"`
	AnonymousID string      `json:"anonymous_id,omitempty"`
	Roles       []string    `json:"roles,omitempty"`
	SellerID    string      `json:"seller_id,omitempty"`
	TenantID    string      `json:"tenant_id,omitempty"`
	Device      DeviceInfo  `json:"device"`
	Network     NetworkInfo `json:"network"`
	Auth        AuthInfo    `json:"auth"`
}

type LogoutSucceededEvent struct {
	EventID    string    `json:"-"`
	TraceID    string    `json:"-"`
	OccurredAt time.Time `json:"-"`
	AccountID  string    `json:"account_id"`
	UserID     string    `json:"user_id,omitempty"`
	SessionID  string    `json:"session_id,omitempty"`
	AllDevices bool      `json:"all_devices"`
	Reason     string    `json:"reason"`
}

type RefreshReuseDetectedEvent struct {
	EventID               string    `json:"-"`
	TraceID               string    `json:"-"`
	OccurredAt            time.Time `json:"-"`
	AccountID             string    `json:"account_id"`
	UserID                string    `json:"user_id,omitempty"`
	SessionID             string    `json:"session_id"`
	TokenID               string    `json:"token_id"`
	IPHash                string    `json:"ip_hash,omitempty"`
	DeviceFingerprintHash string    `json:"device_fingerprint_hash,omitempty"`
	ActionTaken           string    `json:"action_taken"`
}

type DeviceInfo struct {
	DeviceFingerprintHash string `json:"device_fingerprint_hash,omitempty"`
	UserAgent             string `json:"user_agent,omitempty"`
	Channel               string `json:"channel,omitempty"`
	Locale                string `json:"locale,omitempty"`
	Timezone              string `json:"timezone,omitempty"`
}

type NetworkInfo struct {
	IPHash string `json:"ip_hash,omitempty"`
}

type AuthInfo struct {
	Method  string `json:"method"`
	MFAUsed bool   `json:"mfa_used"`
}

func NewEventID() (string, error) {
	raw := make([]byte, 18)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate event id: %w", err)
	}
	return "evt_" + base64.RawURLEncoding.EncodeToString(raw), nil
}

func BuildLoginSucceededOutboxEvent(event LoginSucceededEvent) (domain.OutboxEvent, error) {
	event = normalizeLoginEvent(event)
	if err := validateSessionStart(event.EventID, event.AccountID, event.UserID, event.SessionID, event.OccurredAt); err != nil {
		return domain.OutboxEvent{}, err
	}
	return buildOutboxEvent(
		EventTypeAuthLoginSucceeded,
		RoutingKeyAuthLoginSucceeded,
		AggregateTypeSession,
		event.SessionID,
		event.EventID,
		event.TraceID,
		event.OccurredAt,
		event,
	)
}

func BuildSignupSucceededOutboxEvent(event SignupSucceededEvent) (domain.OutboxEvent, error) {
	event = normalizeSignupEvent(event)
	if err := validateSessionStart(event.EventID, event.AccountID, event.UserID, event.SessionID, event.OccurredAt); err != nil {
		return domain.OutboxEvent{}, err
	}
	return buildOutboxEvent(
		EventTypeAuthSignupSucceeded,
		RoutingKeyAuthSignupSucceeded,
		AggregateTypeSession,
		event.SessionID,
		event.EventID,
		event.TraceID,
		event.OccurredAt,
		event,
	)
}

func BuildLogoutSucceededOutboxEvent(event LogoutSucceededEvent) (domain.OutboxEvent, error) {
	event = normalizeLogoutEvent(event)
	if event.EventID == "" || event.AccountID == "" || event.OccurredAt.IsZero() {
		return domain.OutboxEvent{}, ErrInvalidEvent
	}
	if !event.AllDevices && event.SessionID == "" {
		return domain.OutboxEvent{}, ErrInvalidEvent
	}

	aggregateType := AggregateTypeSession
	aggregateID := event.SessionID
	if event.AllDevices {
		aggregateType = AggregateTypeAccount
		aggregateID = event.AccountID
	}

	return buildOutboxEvent(
		EventTypeAuthLogoutSucceeded,
		RoutingKeyAuthLogoutSucceeded,
		aggregateType,
		aggregateID,
		event.EventID,
		event.TraceID,
		event.OccurredAt,
		event,
	)
}

func BuildRefreshReuseDetectedOutboxEvent(event RefreshReuseDetectedEvent) (domain.OutboxEvent, error) {
	event = normalizeRefreshReuseEvent(event)
	if event.EventID == "" || event.AccountID == "" || event.SessionID == "" || event.TokenID == "" || event.OccurredAt.IsZero() {
		return domain.OutboxEvent{}, ErrInvalidEvent
	}
	return buildOutboxEvent(
		EventTypeRefreshTokenReuseDetected,
		RoutingKeyRefreshTokenReuseDetected,
		AggregateTypeSession,
		event.SessionID,
		event.EventID,
		event.TraceID,
		event.OccurredAt,
		event,
	)
}

func buildOutboxEvent(eventType string, routingKey string, aggregateType string, aggregateID string, eventID string, traceID string, occurredAt time.Time, payload any) (domain.OutboxEvent, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return domain.OutboxEvent{}, fmt.Errorf("marshal session event payload: %w", err)
	}

	envelope := EventEnvelope{
		EventID:       eventID,
		EventType:     eventType,
		Version:       EventVersion,
		OccurredAt:    occurredAt.UTC(),
		Producer:      ProducerAuthService,
		TraceID:       strings.TrimSpace(traceID),
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
		Payload:       payloadBytes,
	}

	envelopeBytes, err := json.Marshal(envelope)
	if err != nil {
		return domain.OutboxEvent{}, fmt.Errorf("marshal session event envelope: %w", err)
	}

	now := occurredAt.UTC()
	return domain.OutboxEvent{
		EventID:       eventID,
		EventType:     eventType,
		Version:       EventVersion,
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
		RoutingKey:    routingKey,
		Payload:       envelopeBytes,
		TraceID:       strings.TrimSpace(traceID),
		Status:        domain.OutboxStatusPending,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

func validateSessionStart(eventID string, accountID string, userID string, sessionID string, occurredAt time.Time) error {
	if eventID == "" || accountID == "" || userID == "" || sessionID == "" || occurredAt.IsZero() {
		return ErrInvalidEvent
	}
	return nil
}
