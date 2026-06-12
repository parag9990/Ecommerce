package events

import (
	"errors"
	"testing"
	"time"

	"ecommerce/backend/services/wishlist-service/internal/domain"
)

func TestAvailabilityFromProductEvent(t *testing.T) {
	tests := []struct {
		name         string
		eventType    string
		availability domain.Availability
		supported    bool
	}{
		{name: "deleted", eventType: ProductDeletedEventType, availability: domain.AvailabilityDeleted, supported: true},
		{name: "out of stock", eventType: ProductOutOfStockEventType, availability: domain.AvailabilityOutOfStock, supported: true},
		{name: "unsupported", eventType: ProductPriceChangedType, availability: domain.AvailabilityUnknown, supported: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			availability, supported := AvailabilityFromProductEvent(tt.eventType)
			if availability != tt.availability || supported != tt.supported {
				t.Fatalf("AvailabilityFromProductEvent() = %q/%v, want %q/%v", availability, supported, tt.availability, tt.supported)
			}
		})
	}
}

func TestDecodeProductAvailabilitySyncInput(t *testing.T) {
	input, ignored, err := DecodeProductAvailabilitySyncInput([]byte(`{
		"event_id": " evt_123 ",
		"event_type": "ProductOutOfStock",
		"version": 1,
		"occurred_at": "2026-05-26T10:30:00Z",
		"producer": "product-service",
		"trace_id": " trace_123 ",
		"payload": {
			"product_id": " prod_123 ",
			"variant_id": " var_1 ",
			"reason": "inventory_zero"
		}
	}`))
	if err != nil {
		t.Fatalf("DecodeProductAvailabilitySyncInput returned error: %v", err)
	}
	if ignored {
		t.Fatal("DecodeProductAvailabilitySyncInput ignored supported event")
	}
	if input.EventID != "evt_123" || input.TraceID != "trace_123" {
		t.Fatalf("input ids = %q/%q, want evt_123/trace_123", input.EventID, input.TraceID)
	}
	if input.ProductID != "prod_123" || input.VariantID != "var_1" {
		t.Fatalf("input product = %q/%q, want prod_123/var_1", input.ProductID, input.VariantID)
	}
	if input.Availability != domain.AvailabilityOutOfStock {
		t.Fatalf("input availability = %q, want out_of_stock", input.Availability)
	}
	wantOccurredAt := time.Date(2026, 5, 26, 10, 30, 0, 0, time.UTC)
	if !input.OccurredAt.Equal(wantOccurredAt) {
		t.Fatalf("OccurredAt = %v, want %v", input.OccurredAt, wantOccurredAt)
	}
}

func TestDecodeProductPriceChangeInput(t *testing.T) {
	input, ignored, err := DecodeProductPriceChangeInput([]byte(`{
		"event_id": " evt_price_123 ",
		"event_type": "ProductPriceChanged",
		"version": 1,
		"occurred_at": "2026-05-26T10:30:00Z",
		"producer": "product-service",
		"trace_id": " trace_123 ",
		"payload": {
			"product_id": " prod_123 ",
			"variant_id": " var_1 ",
			"old_price": {"amount": 349900, "currency": "INR"},
			"new_price": {"amount": 299900, "currency": "inr"},
			"title": " Running Shoes ",
			"image_url": " https://cdn.example.com/prod_123/main.jpg ",
			"product_url": " /products/prod_123 "
		}
	}`))
	if err != nil {
		t.Fatalf("DecodeProductPriceChangeInput returned error: %v", err)
	}
	if ignored {
		t.Fatal("DecodeProductPriceChangeInput ignored supported event")
	}
	if input.EventID != "evt_price_123" || input.TraceID != "trace_123" {
		t.Fatalf("input ids = %q/%q, want evt_price_123/trace_123", input.EventID, input.TraceID)
	}
	if input.ProductID != "prod_123" || input.VariantID != "var_1" {
		t.Fatalf("input product = %q/%q, want prod_123/var_1", input.ProductID, input.VariantID)
	}
	if input.NewPrice.Amount != 299900 || input.NewPrice.Currency != "INR" {
		t.Fatalf("input NewPrice = %#v, want INR 299900", input.NewPrice)
	}
	if input.Title != "Running Shoes" || input.ProductURL != "/products/prod_123" {
		t.Fatalf("input template fields = %q/%q", input.Title, input.ProductURL)
	}
}

func TestDecodeProductPriceChangeInputIgnoresUnsupportedEvent(t *testing.T) {
	_, ignored, err := DecodeProductPriceChangeInput([]byte(`{
		"event_id": "evt_123",
		"event_type": "ProductDeleted",
		"version": 1,
		"occurred_at": "2026-05-26T10:30:00Z",
		"producer": "product-service",
		"payload": {"product_id": "prod_123"}
	}`))
	if err != nil {
		t.Fatalf("DecodeProductPriceChangeInput returned error: %v", err)
	}
	if !ignored {
		t.Fatal("unsupported event was not ignored")
	}
}

func TestDecodeProductPriceChangeInputReturnsPermanentErrors(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "missing event id", body: `{"event_type":"ProductPriceChanged","version":1,"producer":"product-service","payload":{"product_id":"prod_123","new_price":{"amount":299900,"currency":"INR"}}}`},
		{name: "missing product", body: `{"event_id":"evt_1","event_type":"ProductPriceChanged","version":1,"producer":"product-service","payload":{"new_price":{"amount":299900,"currency":"INR"}}}`},
		{name: "bad currency", body: `{"event_id":"evt_1","event_type":"ProductPriceChanged","version":1,"producer":"product-service","payload":{"product_id":"prod_123","new_price":{"amount":299900,"currency":"US"}}}`},
		{name: "negative amount", body: `{"event_id":"evt_1","event_type":"ProductPriceChanged","version":1,"producer":"product-service","payload":{"product_id":"prod_123","new_price":{"amount":-1,"currency":"INR"}}}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ignored, err := DecodeProductPriceChangeInput([]byte(tt.body))
			if ignored {
				t.Fatal("bad event was ignored, want error")
			}
			if err == nil {
				t.Fatal("DecodeProductPriceChangeInput error = nil, want permanent error")
			}
			if !IsPermanentProductEventError(err) {
				t.Fatalf("IsPermanentProductEventError(%v) = false, want true", err)
			}
		})
	}
}

func TestDecodeProductAvailabilitySyncInputIgnoresUnsupportedEvent(t *testing.T) {
	_, ignored, err := DecodeProductAvailabilitySyncInput([]byte(`{
		"event_id": "evt_123",
		"event_type": "ProductPriceChanged",
		"version": 1,
		"occurred_at": "2026-05-26T10:30:00Z",
		"producer": "product-service",
		"payload": {"product_id": "prod_123"}
	}`))
	if err != nil {
		t.Fatalf("DecodeProductAvailabilitySyncInput returned error: %v", err)
	}
	if !ignored {
		t.Fatal("unsupported event was not ignored")
	}
}

func TestDecodeProductAvailabilitySyncInputReturnsPermanentErrors(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "malformed json", body: `{`},
		{name: "unsupported version", body: `{"event_id":"evt_1","event_type":"ProductDeleted","version":2,"producer":"product-service","payload":{"product_id":"prod_123"}}`},
		{name: "unexpected producer", body: `{"event_id":"evt_1","event_type":"ProductDeleted","version":1,"producer":"catalog","payload":{"product_id":"prod_123"}}`},
		{name: "missing product", body: `{"event_id":"evt_1","event_type":"ProductDeleted","version":1,"producer":"product-service","payload":{}}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ignored, err := DecodeProductAvailabilitySyncInput([]byte(tt.body))
			if ignored {
				t.Fatal("bad event was ignored, want error")
			}
			if err == nil {
				t.Fatal("DecodeProductAvailabilitySyncInput error = nil, want permanent error")
			}
			if !IsPermanentProductEventError(err) {
				t.Fatalf("IsPermanentProductEventError(%v) = false, want true", err)
			}
			var eventErr *ProductEventError
			if !errors.As(err, &eventErr) || eventErr.Reason == "" {
				t.Fatalf("error = %#v, want ProductEventError with reason", err)
			}
		})
	}
}
