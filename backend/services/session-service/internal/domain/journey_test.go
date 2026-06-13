package domain

import (
	"errors"
	"testing"
	"time"
)

func TestBuildJourneySummaryCalculatesCountersAndMilestones(t *testing.T) {
	session := validSession()
	now := session.StartedAt
	home := "/"
	product := "/products/prod_123"
	checkout := "/checkout/success"
	events := []SessionEvent{
		journeyEvent("evt_004", EventPaymentResult, checkout, now.Add(8*time.Minute), map[string]any{"status": "success"}),
		journeyEvent("evt_001", EventPageView, home, now, map[string]any{"title": "Home"}),
		journeyEvent("evt_003", EventAddToCart, product, now.Add(4*time.Minute), map[string]any{"product_id": "prod_123"}),
		journeyEvent("evt_002", EventProductView, product, now.Add(3*time.Minute), map[string]any{"product_id": "prod_123"}),
	}

	summary := BuildJourneySummary(session, events, now.Add(9*time.Minute), 2)

	if summary.EntryPage != home || summary.ExitPage != checkout {
		t.Fatalf("unexpected entry/exit pages: %q/%q", summary.EntryPage, summary.ExitPage)
	}
	if summary.DurationSeconds != int64(8*time.Minute/time.Second) {
		t.Fatalf("unexpected duration %d", summary.DurationSeconds)
	}
	if summary.TotalEvents != 4 || summary.ProductsViewed != 1 || summary.CartActions != 1 || !summary.PaymentCompleted {
		t.Fatalf("unexpected counters: %+v", summary)
	}
	if len(summary.Milestones) != 4 {
		t.Fatalf("expected four milestones, got %+v", summary.Milestones)
	}
	if len(summary.TopPaths) != 2 || summary.TopPaths[0].Path != product || summary.TopPaths[0].Count != 2 {
		t.Fatalf("unexpected top paths: %+v", summary.TopPaths)
	}
}

func TestJourneySummaryValidateRejectsNegativeCounts(t *testing.T) {
	session := validSession()
	summary := BuildJourneySummary(session, nil, session.StartedAt, 10)
	summary.TotalEvents = -1

	err := summary.Validate()
	if !errors.Is(err, ErrInvalidJourney) {
		t.Fatalf("expected invalid journey error, got %v", err)
	}
}

func journeyEvent(eventID string, eventType EventType, path string, occurredAt time.Time, properties map[string]any) SessionEvent {
	return SessionEvent{
		EventID:       eventID,
		SessionID:     "sess_01HX9ZK6N5M2V4B7S8T9Q0R1P2",
		AnonymousID:   "anon_01HX9ZJY35F7K2M4N6P8R0T1V2",
		EventType:     eventType,
		Path:          &path,
		Properties:    properties,
		OccurredAt:    occurredAt,
		ReceivedAt:    occurredAt.Add(time.Second),
		SchemaVersion: CurrentSessionEventSchemaVersion,
	}
}
