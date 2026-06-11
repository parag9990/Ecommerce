package provider_test

import (
	"testing"

	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/provider"
)

func TestNormalizeStripeLikeStatus(t *testing.T) {
	tests := map[string]provider.IntentStatus{
		"requires_action":         provider.IntentStatusRequiresAction,
		"requires_confirmation":   provider.IntentStatusRequiresAction,
		"requires_capture":        provider.IntentStatusAuthorized,
		"succeeded":               provider.IntentStatusCaptured,
		"canceled":                provider.IntentStatusFailed,
		"requires_payment_method": provider.IntentStatusInitiated,
		"processing":              provider.IntentStatusInitiated,
	}
	for raw, want := range tests {
		got, ok := provider.NormalizeStripeLikeStatus(raw)
		if !ok || got != want {
			t.Fatalf("NormalizeStripeLikeStatus(%q) = %q/%v, want %q/true", raw, got, ok, want)
		}
	}
}

func TestNormalizeRazorpayLikeStatus(t *testing.T) {
	tests := map[string]provider.IntentStatus{
		"created":    provider.IntentStatusInitiated,
		"attempted":  provider.IntentStatusInitiated,
		"authorized": provider.IntentStatusAuthorized,
		"paid":       provider.IntentStatusCaptured,
		"captured":   provider.IntentStatusCaptured,
		"failed":     provider.IntentStatusFailed,
	}
	for raw, want := range tests {
		got, ok := provider.NormalizeRazorpayLikeStatus(raw)
		if !ok || got != want {
			t.Fatalf("NormalizeRazorpayLikeStatus(%q) = %q/%v, want %q/true", raw, got, ok, want)
		}
	}
}

func TestNormalizeUnknownStatus(t *testing.T) {
	if got, ok := provider.NormalizeStripeLikeStatus("provider_new_state"); ok || got != "" {
		t.Fatalf("NormalizeStripeLikeStatus(unknown) = %q/%v, want empty/false", got, ok)
	}
	if got, ok := provider.NormalizeRazorpayLikeStatus("provider_new_state"); ok || got != "" {
		t.Fatalf("NormalizeRazorpayLikeStatus(unknown) = %q/%v, want empty/false", got, ok)
	}
}
