package events

import (
	"context"
	"fmt"
	"testing"

	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/usecase"
)

func TestInteractionHandlerSendsInvalidJSONToDLQ(t *testing.T) {
	handler, ingestor, dlq := newTestHandler(t)

	if err := handler.Handle(context.Background(), Message{Topic: "recommendation.events", Value: []byte(`{`)}); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if ingestor.calls != 0 {
		t.Fatalf("ingestor calls = %d, want 0", ingestor.calls)
	}
	if len(dlq.messages) != 1 {
		t.Fatalf("dlq messages = %d, want 1", len(dlq.messages))
	}
	if dlq.messages[0].Reason == "" {
		t.Fatal("dlq reason is empty")
	}
}

func TestInteractionHandlerRetriesStorageError(t *testing.T) {
	handler, ingestor, dlq := newTestHandler(t)
	ingestor.errs = []error{
		fmt.Errorf("%w: mongo timeout", domain.ErrRecommendationStorage),
		nil,
	}
	raw := []byte(`{
		"event_id":"evt_view_retry",
		"event_type":"ProductViewed",
		"version":1,
		"occurred_at":"2026-05-27T10:30:00Z",
		"producer":"session-service",
		"payload":{"user_id":"user_123","product_id":"prod_123"}
	}`)

	if err := handler.Handle(context.Background(), Message{Topic: "recommendation.events", Value: raw}); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if ingestor.calls != 2 {
		t.Fatalf("ingestor calls = %d, want 2", ingestor.calls)
	}
	if len(dlq.messages) != 0 {
		t.Fatalf("dlq messages = %d, want 0", len(dlq.messages))
	}
}

func TestInteractionHandlerDLQsAfterRetryExhaustion(t *testing.T) {
	handler, ingestor, dlq := newTestHandler(t)
	ingestor.errs = []error{
		fmt.Errorf("%w: mongo timeout", domain.ErrRecommendationStorage),
		fmt.Errorf("%w: mongo timeout", domain.ErrRecommendationStorage),
	}
	raw := []byte(`{
		"event_id":"evt_view_fail",
		"event_type":"ProductViewed",
		"version":1,
		"occurred_at":"2026-05-27T10:30:00Z",
		"producer":"session-service",
		"payload":{"user_id":"user_123","product_id":"prod_123"}
	}`)

	if err := handler.Handle(context.Background(), Message{Topic: "recommendation.events", Value: raw}); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if ingestor.calls != 2 {
		t.Fatalf("ingestor calls = %d, want 2", ingestor.calls)
	}
	if len(dlq.messages) != 1 {
		t.Fatalf("dlq messages = %d, want 1", len(dlq.messages))
	}
	if dlq.messages[0].Attempts != 2 {
		t.Fatalf("dlq attempts = %d, want 2", dlq.messages[0].Attempts)
	}
}

func TestInteractionHandlerAcceptsFeatureRepairAfterRawDuplicate(t *testing.T) {
	handler, ingestor, dlq := newTestHandler(t)
	ingestor.result = &usecase.InteractionIngestionResult{
		Duplicates:      1,
		FeaturesApplied: 1,
	}
	raw := []byte(`{
		"event_id":"evt_feature_repair",
		"event_type":"ProductViewed",
		"version":1,
		"occurred_at":"2026-05-27T10:30:00Z",
		"producer":"session-service",
		"payload":{"user_id":"user_123","product_id":"prod_123"}
	}`)

	if err := handler.Handle(context.Background(), Message{Topic: "recommendation.events", Value: raw}); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if ingestor.calls != 1 || len(dlq.messages) != 0 {
		t.Fatalf("ingestor calls = %d dlq messages = %d, want repaired event accepted", ingestor.calls, len(dlq.messages))
	}
}

func newTestHandler(t *testing.T) (*InteractionHandler, *fakeInteractionIngestor, *fakeDeadLetterPublisher) {
	t.Helper()
	mapper := newTestMapper(t)
	ingestor := &fakeInteractionIngestor{}
	dlq := &fakeDeadLetterPublisher{}
	handler, err := NewInteractionHandler(
		mapper,
		ingestor,
		dlq,
		nil,
		InteractionHandlerConfig{
			ServiceName:      "recommendation-service",
			Topic:            "recommendation.events",
			MaxRetryAttempts: 2,
			RetryBackoffs:    nil,
		},
		nil,
	)
	if err != nil {
		t.Fatalf("NewInteractionHandler() error = %v", err)
	}
	return handler, ingestor, dlq
}

type fakeInteractionIngestor struct {
	calls  int
	errs   []error
	result *usecase.InteractionIngestionResult
}

func (f *fakeInteractionIngestor) IngestInteractions(ctx context.Context, interactions []domain.UserInteraction) (usecase.InteractionIngestionResult, error) {
	f.calls++
	if len(f.errs) >= f.calls {
		if err := f.errs[f.calls-1]; err != nil {
			return usecase.InteractionIngestionResult{}, err
		}
	}
	if f.result != nil {
		return *f.result, nil
	}
	return usecase.InteractionIngestionResult{Stored: len(interactions)}, nil
}

type fakeDeadLetterPublisher struct {
	messages []DeadLetterMessage
}

func (f *fakeDeadLetterPublisher) Publish(ctx context.Context, message DeadLetterMessage) error {
	f.messages = append(f.messages, message)
	return nil
}
