package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"testing"

	"product-service/internal/domain"
)

func TestProductEventServiceBuildsAndQueuesOutboxEvent(t *testing.T) {
	outbox := &productEventMemoryOutbox{}
	service, err := NewProductEventService(
		outbox,
		&eventSequenceIDs{},
		fixedClock{},
		slog.Default(),
		ProductEventServiceOptions{
			Topic:  "product.events",
			Source: "product-service",
		},
	)
	if err != nil {
		t.Fatalf("NewProductEventService returned error: %v", err)
	}

	product := validDraftProduct()
	product.Status = domain.ProductStatusPublished
	product.UpdatedAt = testNow()
	ctx := WithTraceID(WithRequestID(context.Background(), "req_123"), "trace_456")

	event, err := service.BuildProductEvent(ctx, domain.ProductEventPublished, product, testNow())
	if err != nil {
		t.Fatalf("BuildProductEvent returned error: %v", err)
	}
	if event.ID != "evt_generated" {
		t.Fatalf("event id = %s, want evt_generated", event.ID)
	}
	if event.EventType != string(domain.ProductEventPublished) {
		t.Fatalf("event type = %s, want ProductPublished", event.EventType)
	}
	if event.RequestID != "req_123" || event.TraceID != "trace_456" {
		t.Fatalf("correlation ids = %s/%s", event.RequestID, event.TraceID)
	}
	if event.Payload["search_action"] != string(domain.SearchActionUpsert) {
		t.Fatalf("search action = %v, want UPSERT", event.Payload["search_action"])
	}
	if event.Payload["in_stock"] != true {
		t.Fatalf("in_stock = %v, want true", event.Payload["in_stock"])
	}

	if err := service.QueueProductEvent(ctx, *event); err != nil {
		t.Fatalf("QueueProductEvent returned error: %v", err)
	}
	if len(outbox.events) != 1 || outbox.events[0].ID != "evt_generated" {
		t.Fatalf("queued events = %+v", outbox.events)
	}
}

type eventSequenceIDs struct {
	next int
}

func (g *eventSequenceIDs) NewProductEventID() string {
	g.next++
	if g.next == 1 {
		return "evt_generated"
	}
	return fmt.Sprintf("evt_generated_%d", g.next)
}

type productEventMemoryOutbox struct {
	events []domain.ProductOutboxEvent
}

func (o *productEventMemoryOutbox) InsertOutboxEvent(ctx context.Context, event *domain.ProductOutboxEvent) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	o.events = append(o.events, *event)
	return nil
}
