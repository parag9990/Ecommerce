package domain

import (
	"errors"
	"testing"
	"time"
)

func TestWishlistAnalyticsEventValidation(t *testing.T) {
	now := time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC)
	event := WishlistAnalyticsEvent{
		EventID:   "evt_wish_123",
		EventType: WishlistEventItemAdded,
		Version:   WishlistEventVersion,
		Topic:     "recommendation.events",
		Payload: WishlistAnalyticsPayload{
			UserID:       "user_123",
			ProductID:    "prod_123",
			VariantID:    "var_1",
			Action:       WishlistEventActionAdd,
			Source:       WishlistEventSource,
			Availability: AvailabilityInStock,
			LastKnownPrice: &Money{
				Amount:   299900,
				Currency: "INR",
			},
		},
		Status:      WishlistEventPending,
		NextRetryAt: now,
		OccurredAt:  now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := event.Validate(); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
}

func TestWishlistAnalyticsEventRejectsMismatchedAction(t *testing.T) {
	now := time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC)
	event := WishlistAnalyticsEvent{
		EventID:   "evt_wish_123",
		EventType: WishlistEventItemRemoved,
		Version:   WishlistEventVersion,
		Topic:     "recommendation.events",
		Payload: WishlistAnalyticsPayload{
			UserID:       "user_123",
			ProductID:    "prod_123",
			Action:       WishlistEventActionAdd,
			Source:       WishlistEventSource,
			Availability: AvailabilityInStock,
		},
		Status:      WishlistEventPending,
		NextRetryAt: now,
		OccurredAt:  now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := event.Validate(); !errors.Is(err, ErrInvalidWishlist) {
		t.Fatalf("Validate error = %v, want ErrInvalidWishlist", err)
	}
}
