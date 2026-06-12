package events

import (
	"context"
	"errors"
	"testing"

	"ecommerce/backend/services/wishlist-service/internal/domain"
	"ecommerce/backend/services/wishlist-service/internal/usecase"
)

func TestProductConsumerHandlesOutOfStockEvent(t *testing.T) {
	sync := &fakeAvailabilitySync{}
	consumer := mustProductConsumer(t, sync)

	err := consumer.HandleMessage(context.Background(), []byte(`{
		"event_id": "evt_123",
		"event_type": "ProductOutOfStock",
		"version": 1,
		"occurred_at": "2026-05-26T10:30:00Z",
		"producer": "product-service",
		"trace_id": "trace_123",
		"payload": {
			"product_id": "prod_123",
			"variant_id": "var_1"
		}
	}`))
	if err != nil {
		t.Fatalf("HandleMessage returned error: %v", err)
	}
	if sync.calls != 1 {
		t.Fatalf("sync calls = %d, want 1", sync.calls)
	}
	if sync.input.ProductID != "prod_123" || sync.input.VariantID != "var_1" {
		t.Fatalf("sync product = %q/%q, want prod_123/var_1", sync.input.ProductID, sync.input.VariantID)
	}
	if sync.input.Availability != domain.AvailabilityOutOfStock {
		t.Fatalf("sync availability = %q, want out_of_stock", sync.input.Availability)
	}
}

func TestProductConsumerIgnoresUnsupportedEvent(t *testing.T) {
	sync := &fakeAvailabilitySync{}
	consumer := mustProductConsumer(t, sync)

	err := consumer.HandleMessage(context.Background(), []byte(`{
		"event_id": "evt_123",
		"event_type": "ProductPriceChanged",
		"version": 1,
		"occurred_at": "2026-05-26T10:30:00Z",
		"producer": "product-service",
		"payload": {"product_id": "prod_123"}
	}`))
	if err != nil {
		t.Fatalf("HandleMessage returned error: %v", err)
	}
	if sync.calls != 0 {
		t.Fatalf("sync calls = %d, want 0", sync.calls)
	}
}

func TestProductConsumerHandlesPriceChangedEvent(t *testing.T) {
	priceDrop := &fakePriceChangeUsecase{}
	consumer, err := NewProductConsumer(&fakeAvailabilitySync{}, nil, WithPriceChangeUsecase(priceDrop))
	if err != nil {
		t.Fatalf("NewProductConsumer returned error: %v", err)
	}

	err = consumer.HandleMessage(context.Background(), []byte(`{
		"event_id": "evt_price_123",
		"event_type": "ProductPriceChanged",
		"version": 1,
		"occurred_at": "2026-05-26T10:30:00Z",
		"producer": "product-service",
		"trace_id": "trace_123",
		"payload": {
			"product_id": "prod_123",
			"variant_id": "var_1",
			"new_price": {"amount": 299900, "currency": "INR"},
			"title": "Running Shoes"
		}
	}`))
	if err != nil {
		t.Fatalf("HandleMessage returned error: %v", err)
	}
	if priceDrop.calls != 1 {
		t.Fatalf("priceDrop calls = %d, want 1", priceDrop.calls)
	}
	if priceDrop.input.ProductID != "prod_123" || priceDrop.input.VariantID != "var_1" {
		t.Fatalf("priceDrop product = %q/%q, want prod_123/var_1", priceDrop.input.ProductID, priceDrop.input.VariantID)
	}
	if priceDrop.input.NewPrice.Amount != 299900 || priceDrop.input.NewPrice.Currency != "INR" {
		t.Fatalf("priceDrop new price = %#v, want INR 299900", priceDrop.input.NewPrice)
	}
}

func TestProductConsumerReturnsPermanentErrorForBadEvent(t *testing.T) {
	consumer := mustProductConsumer(t, &fakeAvailabilitySync{})

	err := consumer.HandleMessage(context.Background(), []byte(`{"event_id":"evt_1","event_type":"ProductDeleted","version":1,"producer":"product-service","payload":{}}`))
	if err == nil {
		t.Fatal("HandleMessage error = nil, want permanent error")
	}
	if !IsPermanentProductEventError(err) {
		t.Fatalf("IsPermanentProductEventError(%v) = false, want true", err)
	}
}

func TestProductConsumerReturnsPermanentErrorForBadPriceEvent(t *testing.T) {
	consumer, err := NewProductConsumer(&fakeAvailabilitySync{}, nil, WithPriceChangeUsecase(&fakePriceChangeUsecase{}))
	if err != nil {
		t.Fatalf("NewProductConsumer returned error: %v", err)
	}

	err = consumer.HandleMessage(context.Background(), []byte(`{
		"event_id": "evt_price_123",
		"event_type": "ProductPriceChanged",
		"version": 1,
		"producer": "product-service",
		"payload": {"product_id":"prod_123","new_price":{"amount":299900,"currency":"US"}}
	}`))
	if err == nil {
		t.Fatal("HandleMessage error = nil, want permanent error")
	}
	if !IsPermanentProductEventError(err) {
		t.Fatalf("IsPermanentProductEventError(%v) = false, want true", err)
	}
}

func TestProductConsumerReturnsRetryableUsecaseError(t *testing.T) {
	syncErr := errors.New("mongo timeout")
	consumer := mustProductConsumer(t, &fakeAvailabilitySync{err: syncErr})

	err := consumer.HandleMessage(context.Background(), []byte(`{
		"event_id": "evt_123",
		"event_type": "ProductDeleted",
		"version": 1,
		"occurred_at": "2026-05-26T10:30:00Z",
		"producer": "product-service",
		"payload": {"product_id": "prod_123"}
	}`))
	if !errors.Is(err, syncErr) {
		t.Fatalf("HandleMessage error = %v, want %v", err, syncErr)
	}
	if IsPermanentProductEventError(err) {
		t.Fatalf("usecase error = %v was marked permanent", err)
	}
}

func TestProductConsumerReturnsRetryablePriceUsecaseError(t *testing.T) {
	priceErr := errors.New("notification topic unavailable")
	consumer, err := NewProductConsumer(&fakeAvailabilitySync{}, nil, WithPriceChangeUsecase(&fakePriceChangeUsecase{err: priceErr}))
	if err != nil {
		t.Fatalf("NewProductConsumer returned error: %v", err)
	}

	err = consumer.HandleMessage(context.Background(), []byte(`{
		"event_id": "evt_price_123",
		"event_type": "ProductPriceChanged",
		"version": 1,
		"producer": "product-service",
		"payload": {"product_id":"prod_123","new_price":{"amount":299900,"currency":"INR"}}
	}`))
	if !errors.Is(err, priceErr) {
		t.Fatalf("HandleMessage error = %v, want %v", err, priceErr)
	}
	if IsPermanentProductEventError(err) {
		t.Fatalf("price usecase error = %v was marked permanent", err)
	}
}

func mustProductConsumer(t *testing.T, sync AvailabilitySyncUsecase) *ProductConsumer {
	t.Helper()
	consumer, err := NewProductConsumer(sync, nil)
	if err != nil {
		t.Fatalf("NewProductConsumer returned error: %v", err)
	}
	return consumer
}

type fakeAvailabilitySync struct {
	input usecase.SyncProductAvailabilityInput
	err   error
	calls int
}

type fakePriceChangeUsecase struct {
	input usecase.PriceChangeInput
	err   error
	calls int
}

func (f *fakePriceChangeUsecase) HandleProductPriceChanged(ctx context.Context, input usecase.PriceChangeInput) (usecase.PriceDropResult, error) {
	f.calls++
	f.input = input
	if f.err != nil {
		return usecase.PriceDropResult{}, f.err
	}
	return usecase.PriceDropResult{
		ProductID:             input.ProductID,
		VariantID:             input.VariantID,
		NewPrice:              input.NewPrice,
		CandidateCount:        1,
		NotificationCount:     1,
		SnapshotModifiedCount: 1,
	}, nil
}

func (f *fakeAvailabilitySync) SyncProductAvailability(ctx context.Context, input usecase.SyncProductAvailabilityInput) (usecase.AvailabilitySyncResult, error) {
	f.calls++
	f.input = input
	if f.err != nil {
		return usecase.AvailabilitySyncResult{}, f.err
	}
	return usecase.AvailabilitySyncResult{
		ProductID:     input.ProductID,
		VariantID:     input.VariantID,
		Availability:  input.Availability,
		MatchedCount:  1,
		ModifiedCount: 1,
	}, nil
}
