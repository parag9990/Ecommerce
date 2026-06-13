package indexer

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
	searchevents "github.com/example/ecommerce-platform/backend/services/search-service/internal/events"
)

func TestProductIndexerUpsertsPublishedProduct(t *testing.T) {
	repo := &fakeProductIndexRepository{}
	idx := newTestIndexer(t, repo)
	envelope := validEnvelope(domain.ProductEventPublished)
	envelope.Payload.CategoryIDs = []string{"cat_shoes", " cat_running ", "cat_shoes", ""}
	envelope.Payload.Description = " Lightweight "
	envelope.Payload.Brand = " Nike "

	result, err := idx.Handle(context.Background(), envelope)
	if err != nil {
		t.Fatalf("handle: %v", err)
	}

	if result.Action != ActionUpsert {
		t.Fatalf("action = %q", result.Action)
	}
	if len(repo.upserts) != 1 {
		t.Fatalf("upsert count = %d", len(repo.upserts))
	}
	doc := repo.upserts[0]
	if doc.ID != "prod_123" || doc.Title != "Running Shoes" {
		t.Fatalf("unexpected doc: %#v", doc)
	}
	if len(doc.CategoryIDs) != 2 || doc.CategoryIDs[0] != "cat_shoes" || doc.CategoryIDs[1] != "cat_running" {
		t.Fatalf("category ids not normalized: %#v", doc.CategoryIDs)
	}
	if doc.Description == nil || *doc.Description != "Lightweight" {
		t.Fatalf("description not normalized: %#v", doc.Description)
	}
	if doc.Brand == nil || *doc.Brand != "Nike" {
		t.Fatalf("brand not normalized: %#v", doc.Brand)
	}
	if doc.CreatedAt != envelope.Payload.UpdatedAt.Unix() {
		t.Fatalf("created_at = %d, want %d", doc.CreatedAt, envelope.Payload.UpdatedAt.Unix())
	}
}

func TestProductIndexerDeletesNonSearchableUpdatedProduct(t *testing.T) {
	repo := &fakeProductIndexRepository{}
	idx := newTestIndexer(t, repo)
	envelope := validEnvelope(domain.ProductEventUpdated)
	envelope.Payload.Status = "draft"
	envelope.Payload.Title = ""
	envelope.Payload.CategoryIDs = nil

	result, err := idx.Handle(context.Background(), envelope)
	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	if result.Action != ActionDelete {
		t.Fatalf("action = %q", result.Action)
	}
	if len(repo.deletes) != 1 || repo.deletes[0] != "prod_123" {
		t.Fatalf("deletes = %#v", repo.deletes)
	}
	if len(repo.upserts) != 0 {
		t.Fatalf("unexpected upserts = %#v", repo.upserts)
	}
}

func TestProductIndexerDeletesProductDeletedEventWithMinimalPayload(t *testing.T) {
	repo := &fakeProductIndexRepository{}
	idx := newTestIndexer(t, repo)
	envelope := validEnvelope(domain.ProductEventDeleted)
	envelope.Payload = searchevents.ProductIndexPayload{
		ProductID: "prod_deleted",
		Status:    "deleted",
		IsDeleted: true,
		UpdatedAt: time.Date(2026, 5, 23, 11, 0, 0, 0, time.UTC),
	}

	result, err := idx.Handle(context.Background(), envelope)
	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	if result.Action != ActionDelete {
		t.Fatalf("action = %q", result.Action)
	}
	if len(repo.deletes) != 1 || repo.deletes[0] != "prod_deleted" {
		t.Fatalf("deletes = %#v", repo.deletes)
	}
}

func TestProductIndexerRejectsInvalidPublishedPayload(t *testing.T) {
	repo := &fakeProductIndexRepository{}
	idx := newTestIndexer(t, repo)
	envelope := validEnvelope(domain.ProductEventPublished)
	badRating := 6.0
	envelope.Payload.Rating = &badRating

	_, err := idx.Handle(context.Background(), envelope)
	if !errors.Is(err, domain.ErrInvalidProductEvent) {
		t.Fatalf("expected ErrInvalidProductEvent, got %v", err)
	}
	if len(repo.upserts) != 0 || len(repo.deletes) != 0 {
		t.Fatalf("unexpected writes: upserts=%#v deletes=%#v", repo.upserts, repo.deletes)
	}
}

func TestProductIndexerRejectsMissingProductID(t *testing.T) {
	repo := &fakeProductIndexRepository{}
	idx := newTestIndexer(t, repo)
	envelope := validEnvelope(domain.ProductEventPublished)
	envelope.Payload.ProductID = ""

	_, err := idx.Handle(context.Background(), envelope)
	if !errors.Is(err, domain.ErrInvalidProductEvent) {
		t.Fatalf("expected ErrInvalidProductEvent, got %v", err)
	}
}

func TestMapProductToDocumentUsesCreatedAtWhenProvided(t *testing.T) {
	createdAt := time.Date(2026, 5, 20, 9, 0, 0, 0, time.UTC)
	envelope := validEnvelope(domain.ProductEventPublished)
	envelope.Payload.CreatedAt = &createdAt

	doc := MapProductToDocument(envelope.Payload, envelope.OccurredAt)
	if doc.CreatedAt != createdAt.Unix() {
		t.Fatalf("created_at = %d, want %d", doc.CreatedAt, createdAt.Unix())
	}
}

type fakeProductIndexRepository struct {
	upserts []domain.ProductDocument
	deletes []string
}

func (r *fakeProductIndexRepository) UpsertProduct(ctx context.Context, doc domain.ProductDocument) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.upserts = append(r.upserts, doc)
	return nil
}

func (r *fakeProductIndexRepository) DeleteProduct(ctx context.Context, productID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.deletes = append(r.deletes, productID)
	return nil
}

func newTestIndexer(t *testing.T, repo ProductIndexRepository) *ProductIndexer {
	t.Helper()
	idx, err := NewProductIndexer(repo, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("new indexer: %v", err)
	}
	return idx
}

func validEnvelope(eventType string) searchevents.Envelope[searchevents.ProductIndexPayload] {
	price := 2499.0
	rating := 4.5
	popularity := int32(982)
	inStock := true
	now := time.Date(2026, 5, 23, 10, 30, 0, 0, time.UTC)
	return searchevents.Envelope[searchevents.ProductIndexPayload]{
		EventID:       "evt_123",
		EventType:     eventType,
		Version:       searchevents.SupportedProductEventVersion,
		Producer:      "product-service",
		RequestID:     "req_123",
		TraceID:       "trace_123",
		CorrelationID: "prod_123",
		OccurredAt:    now,
		Payload: searchevents.ProductIndexPayload{
			ProductID:       "prod_123",
			Title:           "Running Shoes",
			Description:     "Lightweight running shoes",
			Brand:           "Nike",
			CategoryIDs:     []string{"cat_shoes"},
			SellerID:        "seller_456",
			Price:           &price,
			Rating:          &rating,
			PopularityScore: &popularity,
			InStock:         &inStock,
			Status:          domain.ProductStatusPublished,
			UpdatedAt:       now.Add(-time.Minute),
		},
	}
}
