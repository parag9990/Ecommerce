package retry

import (
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/provider"
)

func TestDefaultPolicyStopsAfterFourthAttempt(t *testing.T) {
	t.Parallel()

	policy := DefaultPolicy()
	cases := []struct {
		attempt int
		want    time.Duration
		retry   bool
	}{
		{1, 30 * time.Second, true},
		{2, 2 * time.Minute, true},
		{3, 8 * time.Minute, true},
		{4, 0, false},
	}
	for _, test := range cases {
		delay, ok := policy.NextDelay(test.attempt)
		if delay != test.want || ok != test.retry {
			t.Fatalf("NextDelay(%d) = (%s, %t), want (%s, %t)",
				test.attempt, delay, ok, test.want, test.retry)
		}
	}
}

func TestNewPolicyAndFailureClassification(t *testing.T) {
	t.Parallel()

	if _, err := NewPolicy(4, []time.Duration{time.Second, time.Second, 3 * time.Second}); err == nil {
		t.Fatal("NewPolicy() accepted non-increasing delays")
	}
	if got := ClassifyError(provider.ErrProviderUnavailable); got != FailureTransient {
		t.Fatalf("ClassifyError(unavailable) = %q", got)
	}
	if got := ClassifyError(provider.ErrProviderRejected); got != FailureTerminal {
		t.Fatalf("ClassifyError(rejected) = %q", got)
	}
	if got := ClassifyError(domain.ErrUnsupportedTemplateKey); got != FailureTerminal {
		t.Fatalf("ClassifyError(unclassified template) = %q", got)
	}
}
