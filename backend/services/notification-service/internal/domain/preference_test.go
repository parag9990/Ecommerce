package domain

import (
	"errors"
	"testing"
	"time"
)

func TestDefaultPreferenceIsSafeAndValid(t *testing.T) {
	t.Parallel()

	preference := DefaultPreference("user_1", time.Now())
	if err := preference.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if !preference.EmailEnabled || preference.SMSEnabled || preference.PushEnabled || preference.MarketingEnabled {
		t.Fatalf("default preference = %+v", preference)
	}
}

func TestPreferenceDecideEnforcesPurposeAndChannelConsent(t *testing.T) {
	t.Parallel()

	preference := DefaultPreference("user_1", time.Now())
	preference.MarketingEnabled = true
	tests := []struct {
		name    string
		purpose NotificationPurpose
		channel Channel
		allowed bool
		reason  ConsentReason
	}{
		{name: "transactional email default", purpose: PurposeTransactional, channel: ChannelEmail, allowed: true, reason: ConsentReasonTransactionalChannelAllowed},
		{name: "transactional SMS disabled", purpose: PurposeTransactional, channel: ChannelSMS, reason: ConsentReasonChannelOptedOut},
		{name: "marketing email consent", purpose: PurposeMarketing, channel: ChannelEmail, allowed: true, reason: ConsentReasonMarketingConsentPresent},
		{name: "WhatsApp consent unavailable", purpose: PurposeMarketing, channel: ChannelWhatsAppLike, reason: ConsentReasonChannelConsentUnavailable},
		{name: "security SMS bypasses settings", purpose: PurposeSecurity, channel: ChannelSMS, allowed: true, reason: ConsentReasonSecurityRequested},
		{name: "security push rejected", purpose: PurposeSecurity, channel: ChannelPush, reason: ConsentReasonSecurityChannelNotAllowed},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			got, err := preference.Decide(test.purpose, test.channel)
			if err != nil || got.Allowed != test.allowed || got.Reason != test.reason {
				t.Fatalf("Decide() = %+v, %v", got, err)
			}
		})
	}

	preference.MarketingEnabled = false
	got, err := preference.Decide(PurposeMarketing, ChannelEmail)
	if err != nil || got.Allowed || got.Reason != ConsentReasonMarketingOptedOut {
		t.Fatalf("Decide(marketing opt-out) = %+v, %v", got, err)
	}
}

func TestPreferenceRejectsInvalidRecord(t *testing.T) {
	t.Parallel()

	if err := (Preference{}).Validate(); !errors.Is(err, ErrInvalidPreference) {
		t.Fatalf("Validate() error = %v", err)
	}
}
