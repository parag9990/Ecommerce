package domain

import (
	"fmt"
	"strings"
	"time"
)

// NotificationPurpose determines whether explicit promotional consent is needed.
type NotificationPurpose string

const (
	PurposeSecurity      NotificationPurpose = "security"
	PurposeTransactional NotificationPurpose = "transactional"
	PurposeMarketing     NotificationPurpose = "marketing"
)

type ConsentSource string

const (
	ConsentSourceProfileSettings ConsentSource = "profile_settings"
	ConsentSourceOnboarding      ConsentSource = "onboarding"
	ConsentSourceSupportRequest  ConsentSource = "support_request"
)

type ConsentReason string

const (
	ConsentReasonSecurityRequested           ConsentReason = "security_requested"
	ConsentReasonSecurityChannelNotAllowed   ConsentReason = "security_channel_not_allowed"
	ConsentReasonTransactionalChannelAllowed ConsentReason = "transactional_channel_allowed"
	ConsentReasonMarketingConsentPresent     ConsentReason = "marketing_consent_present"
	ConsentReasonMarketingOptedOut           ConsentReason = "marketing_opted_out"
	ConsentReasonChannelOptedOut             ConsentReason = "channel_opted_out"
	ConsentReasonChannelConsentUnavailable   ConsentReason = "channel_consent_unavailable"
)

func (r ConsentReason) IsSuppressionReason() bool {
	switch r {
	case ConsentReasonSecurityChannelNotAllowed, ConsentReasonMarketingOptedOut,
		ConsentReasonChannelOptedOut, ConsentReasonChannelConsentUnavailable:
		return true
	default:
		return false
	}
}

// Preference is the service-owned user consent projection used at send time.
type Preference struct {
	ID               string        `bson:"_id" json:"id"`
	UserID           string        `bson:"user_id" json:"user_id"`
	EmailEnabled     bool          `bson:"email_enabled" json:"email_enabled"`
	SMSEnabled       bool          `bson:"sms_enabled" json:"sms_enabled"`
	PushEnabled      bool          `bson:"push_enabled" json:"push_enabled"`
	MarketingEnabled bool          `bson:"marketing_enabled" json:"marketing_enabled"`
	ConsentSource    ConsentSource `bson:"consent_source" json:"consent_source"`
	CreatedAt        time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time     `bson:"updated_at" json:"updated_at"`
}

func DefaultPreference(userID string, now time.Time) Preference {
	userID = strings.TrimSpace(userID)
	now = now.UTC()
	return Preference{
		ID:            "pref_" + userID,
		UserID:        userID,
		EmailEnabled:  true,
		ConsentSource: ConsentSourceProfileSettings,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func (p Preference) Validate() error {
	for value, field := range map[string]string{p.ID: "id", p.UserID: "user_id"} {
		if err := requiredText(value, field); err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidPreference, err)
		}
	}
	if !p.ConsentSource.IsSupported() {
		return fmt.Errorf("%w: unsupported consent source %q", ErrInvalidPreference, p.ConsentSource)
	}
	if err := validateRecordTimestamps(p.CreatedAt, p.UpdatedAt); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidPreference, err)
	}
	return nil
}

func (s ConsentSource) IsSupported() bool {
	switch s {
	case ConsentSourceProfileSettings, ConsentSourceOnboarding, ConsentSourceSupportRequest:
		return true
	default:
		return false
	}
}

type ConsentDecision struct {
	Allowed bool
	Reason  ConsentReason
	Purpose NotificationPurpose
}

func (p Preference) Decide(purpose NotificationPurpose, channel Channel) (ConsentDecision, error) {
	if purpose == PurposeSecurity {
		switch channel {
		case ChannelEmail, ChannelSMS:
			return ConsentDecision{Allowed: true, Reason: ConsentReasonSecurityRequested, Purpose: purpose}, nil
		default:
			return ConsentDecision{Reason: ConsentReasonSecurityChannelNotAllowed, Purpose: purpose}, nil
		}
	}
	if purpose != PurposeTransactional && purpose != PurposeMarketing {
		return ConsentDecision{}, fmt.Errorf("%w: unsupported notification purpose %q", ErrInvalidPreference, purpose)
	}
	if channel == ChannelWhatsAppLike {
		return ConsentDecision{Reason: ConsentReasonChannelConsentUnavailable, Purpose: purpose}, nil
	}
	channelEnabled, err := p.channelEnabled(channel)
	if err != nil {
		return ConsentDecision{}, err
	}
	if !channelEnabled {
		return ConsentDecision{Reason: ConsentReasonChannelOptedOut, Purpose: purpose}, nil
	}
	if purpose == PurposeMarketing {
		if !p.MarketingEnabled {
			return ConsentDecision{Reason: ConsentReasonMarketingOptedOut, Purpose: purpose}, nil
		}
		return ConsentDecision{Allowed: true, Reason: ConsentReasonMarketingConsentPresent, Purpose: purpose}, nil
	}
	return ConsentDecision{Allowed: true, Reason: ConsentReasonTransactionalChannelAllowed, Purpose: purpose}, nil
}

func (p Preference) channelEnabled(channel Channel) (bool, error) {
	switch channel {
	case ChannelEmail:
		return p.EmailEnabled, nil
	case ChannelSMS:
		return p.SMSEnabled, nil
	case ChannelPush:
		return p.PushEnabled, nil
	case ChannelWhatsAppLike:
		return false, nil
	default:
		return false, fmt.Errorf("%w: %q", ErrUnsupportedChannel, channel)
	}
}

type DeliverySuppressedError struct {
	Reason  ConsentReason
	Purpose NotificationPurpose
}

func (e *DeliverySuppressedError) Error() string {
	return fmt.Sprintf("%s: %s", ErrDeliverySuppressed, e.Reason)
}

func (e *DeliverySuppressedError) Unwrap() error {
	return ErrDeliverySuppressed
}
