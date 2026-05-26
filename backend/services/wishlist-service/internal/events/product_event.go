package events

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"ecommerce/backend/services/wishlist-service/internal/domain"
	"ecommerce/backend/services/wishlist-service/internal/usecase"
)

const (
	ProductEventVersion        = 1
	ProductEventProducer       = "product-service"
	ProductDeletedEventType    = "ProductDeleted"
	ProductOutOfStockEventType = "ProductOutOfStock"
	ProductPriceChangedType    = "ProductPriceChanged"
)

type ProductEventEnvelope struct {
	EventID    string          `json:"event_id"`
	EventType  string          `json:"event_type"`
	Version    int             `json:"version"`
	OccurredAt time.Time       `json:"occurred_at"`
	Producer   string          `json:"producer"`
	TraceID    string          `json:"trace_id,omitempty"`
	Payload    json.RawMessage `json:"payload"`
}

type ProductAvailabilityPayload struct {
	ProductID string `json:"product_id"`
	VariantID string `json:"variant_id,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

type ProductPriceChangedPayload struct {
	ProductID  string        `json:"product_id"`
	VariantID  string        `json:"variant_id,omitempty"`
	OldPrice   *domain.Money `json:"old_price,omitempty"`
	NewPrice   domain.Money  `json:"new_price"`
	Title      string        `json:"title,omitempty"`
	ImageURL   string        `json:"image_url,omitempty"`
	ProductURL string        `json:"product_url,omitempty"`
}

type ProductEventError struct {
	Reason    string
	Permanent bool
	Err       error
}

func (e *ProductEventError) Error() string {
	if e == nil {
		return ""
	}
	if e.Err == nil {
		return e.Reason
	}
	if e.Reason == "" {
		return e.Err.Error()
	}
	return fmt.Sprintf("%s: %v", e.Reason, e.Err)
}

func (e *ProductEventError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func IsPermanentProductEventError(err error) bool {
	var eventErr *ProductEventError
	return errors.As(err, &eventErr) && eventErr.Permanent
}

func ProductEventErrorReason(err error) string {
	var eventErr *ProductEventError
	if errors.As(err, &eventErr) && eventErr.Reason != "" {
		return eventErr.Reason
	}
	return "processing_failed"
}

func AvailabilityFromProductEvent(eventType string) (domain.Availability, bool) {
	switch strings.TrimSpace(eventType) {
	case ProductDeletedEventType:
		return domain.AvailabilityDeleted, true
	case ProductOutOfStockEventType:
		return domain.AvailabilityOutOfStock, true
	default:
		return domain.AvailabilityUnknown, false
	}
}

func DecodeProductAvailabilitySyncInput(body []byte) (usecase.SyncProductAvailabilityInput, bool, error) {
	var envelope ProductEventEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		return usecase.SyncProductAvailabilityInput{}, false, permanentEventError("decode_envelope_failed", err)
	}

	envelope.EventID = strings.TrimSpace(envelope.EventID)
	envelope.EventType = strings.TrimSpace(envelope.EventType)
	envelope.Producer = strings.TrimSpace(envelope.Producer)
	envelope.TraceID = strings.TrimSpace(envelope.TraceID)

	if envelope.Version != ProductEventVersion {
		return usecase.SyncProductAvailabilityInput{}, false, permanentEventError("unsupported_version", fmt.Errorf("version %d", envelope.Version))
	}
	if envelope.Producer != ProductEventProducer {
		return usecase.SyncProductAvailabilityInput{}, false, permanentEventError("unexpected_producer", fmt.Errorf("producer %q", envelope.Producer))
	}
	if envelope.EventType == "" {
		return usecase.SyncProductAvailabilityInput{}, false, permanentEventError("missing_event_type", errors.New("event_type is required"))
	}

	availability, supported := AvailabilityFromProductEvent(envelope.EventType)
	if !supported {
		return usecase.SyncProductAvailabilityInput{}, true, nil
	}
	if envelope.EventID == "" {
		return usecase.SyncProductAvailabilityInput{}, false, permanentEventError("missing_event_id", errors.New("event_id is required"))
	}
	if len(envelope.Payload) == 0 || strings.EqualFold(strings.TrimSpace(string(envelope.Payload)), "null") {
		return usecase.SyncProductAvailabilityInput{}, false, permanentEventError("missing_payload", errors.New("payload is required"))
	}

	var payload ProductAvailabilityPayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return usecase.SyncProductAvailabilityInput{}, false, permanentEventError("decode_payload_failed", err)
	}
	payload.ProductID = strings.TrimSpace(payload.ProductID)
	payload.VariantID = strings.TrimSpace(payload.VariantID)
	if payload.ProductID == "" {
		return usecase.SyncProductAvailabilityInput{}, false, permanentEventError("missing_product_id", errors.New("payload.product_id is required"))
	}

	return usecase.SyncProductAvailabilityInput{
		EventID:      envelope.EventID,
		EventType:    envelope.EventType,
		ProductID:    payload.ProductID,
		VariantID:    payload.VariantID,
		Availability: availability,
		OccurredAt:   envelope.OccurredAt,
		TraceID:      envelope.TraceID,
	}, false, nil
}

func DecodeProductPriceChangeInput(body []byte) (usecase.PriceChangeInput, bool, error) {
	var envelope ProductEventEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		return usecase.PriceChangeInput{}, false, permanentEventError("decode_envelope_failed", err)
	}

	envelope.EventID = strings.TrimSpace(envelope.EventID)
	envelope.EventType = strings.TrimSpace(envelope.EventType)
	envelope.Producer = strings.TrimSpace(envelope.Producer)
	envelope.TraceID = strings.TrimSpace(envelope.TraceID)

	if envelope.Version != ProductEventVersion {
		return usecase.PriceChangeInput{}, false, permanentEventError("unsupported_version", fmt.Errorf("version %d", envelope.Version))
	}
	if envelope.Producer != ProductEventProducer {
		return usecase.PriceChangeInput{}, false, permanentEventError("unexpected_producer", fmt.Errorf("producer %q", envelope.Producer))
	}
	if envelope.EventType == "" {
		return usecase.PriceChangeInput{}, false, permanentEventError("missing_event_type", errors.New("event_type is required"))
	}
	if envelope.EventType != ProductPriceChangedType {
		return usecase.PriceChangeInput{}, true, nil
	}
	if envelope.EventID == "" {
		return usecase.PriceChangeInput{}, false, permanentEventError("missing_event_id", errors.New("event_id is required"))
	}
	if len(envelope.Payload) == 0 || strings.EqualFold(strings.TrimSpace(string(envelope.Payload)), "null") {
		return usecase.PriceChangeInput{}, false, permanentEventError("missing_payload", errors.New("payload is required"))
	}

	var payload ProductPriceChangedPayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return usecase.PriceChangeInput{}, false, permanentEventError("decode_payload_failed", err)
	}
	payload.ProductID = strings.TrimSpace(payload.ProductID)
	payload.VariantID = strings.TrimSpace(payload.VariantID)
	payload.Title = strings.TrimSpace(payload.Title)
	payload.ImageURL = strings.TrimSpace(payload.ImageURL)
	payload.ProductURL = strings.TrimSpace(payload.ProductURL)
	payload.NewPrice.Currency = strings.ToUpper(strings.TrimSpace(payload.NewPrice.Currency))
	if payload.ProductID == "" {
		return usecase.PriceChangeInput{}, false, permanentEventError("missing_product_id", errors.New("payload.product_id is required"))
	}
	if payload.NewPrice.Amount < 0 {
		return usecase.PriceChangeInput{}, false, permanentEventError("invalid_new_price", errors.New("payload.new_price.amount must be greater than or equal to zero"))
	}
	if err := payload.NewPrice.Validate(); err != nil {
		return usecase.PriceChangeInput{}, false, permanentEventError("invalid_new_price", err)
	}

	return usecase.PriceChangeInput{
		EventID:    envelope.EventID,
		EventType:  envelope.EventType,
		ProductID:  payload.ProductID,
		VariantID:  payload.VariantID,
		NewPrice:   payload.NewPrice,
		Title:      payload.Title,
		ImageURL:   payload.ImageURL,
		ProductURL: payload.ProductURL,
		OccurredAt: envelope.OccurredAt,
		TraceID:    envelope.TraceID,
	}, false, nil
}

func permanentEventError(reason string, err error) *ProductEventError {
	return &ProductEventError{
		Reason:    reason,
		Permanent: true,
		Err:       err,
	}
}
