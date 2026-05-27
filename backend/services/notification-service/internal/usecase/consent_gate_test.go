package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/usecase"
)

type preferenceReader struct {
	preference domain.Preference
	err        error
	calls      int
}

func (r *preferenceReader) Get(context.Context, string) (domain.Preference, error) {
	r.calls++
	return r.preference, r.err
}

func TestConsentGateRequiresLatestMarketingConsent(t *testing.T) {
	t.Parallel()

	reader := &preferenceReader{preference: domain.DefaultPreference("user_1", time.Now())}
	gate := newConsentGate(t, reader)
	decision, err := gate.Evaluate(context.Background(), "user_1", domain.ChannelEmail,
		domain.TemplatePromotionalOffer, false)
	if err != nil || decision.Allowed || decision.Reason != domain.ConsentReasonMarketingOptedOut ||
		reader.calls != 1 {
		t.Fatalf("Evaluate() = %+v, %v; reads = %d", decision, err, reader.calls)
	}
}

func TestConsentGateRestrictsSecurityBypassToTrustedOTP(t *testing.T) {
	t.Parallel()

	reader := &preferenceReader{err: errors.New("should not read preference")}
	gate := newConsentGate(t, reader)
	if _, err := gate.Evaluate(context.Background(), "user_1", domain.ChannelEmail,
		domain.TemplateOTPVerification, false); !errors.Is(err, domain.ErrInvalidPreference) {
		t.Fatalf("Evaluate(untrusted OTP) error = %v", err)
	}
	decision, err := gate.Evaluate(context.Background(), "user_1", domain.ChannelEmail,
		domain.TemplateOTPVerification, true)
	if err != nil || !decision.Allowed || reader.calls != 0 {
		t.Fatalf("Evaluate(trusted OTP) = %+v, %v; reads = %d", decision, err, reader.calls)
	}
}

func newConsentGate(t *testing.T, reader usecase.PreferenceReader) *usecase.ConsentGate {
	t.Helper()
	gate, err := usecase.NewConsentGate(reader)
	if err != nil {
		t.Fatalf("NewConsentGate() error = %v", err)
	}
	return gate
}

func allowedConsentGate(t *testing.T) *usecase.ConsentGate {
	t.Helper()
	preference := domain.DefaultPreference("user_1", time.Now())
	preference.EmailEnabled = true
	preference.MarketingEnabled = true
	return newConsentGate(t, &preferenceReader{preference: preference})
}
