package domain

import (
	"errors"
	"testing"
	"time"
)

func TestSourceEventTypeMapping(t *testing.T) {
	tests := []struct {
		eventType  SourceEventType
		wantType   InteractionType
		wantWeight int
	}{
		{EventProductViewed, InteractionProductView, 1},
		{EventCartItemAdded, InteractionAddToCart, 4},
		{EventWishlistItemAdded, InteractionWishlistAdd, 3},
		{EventWishlistItemRemoved, InteractionWishlistRemove, -2},
		{EventOrderPaid, InteractionPurchase, 8},
		{EventPurchaseCompleted, InteractionPurchase, 8},
	}

	for _, tt := range tests {
		t.Run(string(tt.eventType), func(t *testing.T) {
			gotType, ok := tt.eventType.NormalizedInteractionType()
			if !ok || gotType != tt.wantType {
				t.Fatalf("NormalizedInteractionType() = %s/%t, want %s/true", gotType, ok, tt.wantType)
			}
			gotWeight, ok := tt.eventType.InteractionWeight()
			if !ok || gotWeight != tt.wantWeight {
				t.Fatalf("InteractionWeight() = %d/%t, want %d/true", gotWeight, ok, tt.wantWeight)
			}
		})
	}
}

func TestUserInteractionValidateRequiresIdentity(t *testing.T) {
	interaction := validInteraction()
	interaction.UserID = ""
	interaction.AnonymousID = ""

	if err := interaction.Validate(); !errors.Is(err, ErrInvalidInteraction) {
		t.Fatalf("Validate() error = %v, want %v", err, ErrInvalidInteraction)
	}
}

func TestInteractionDedupeKeyAndIDAreStable(t *testing.T) {
	key := NewInteractionDedupeKey("evt_123", InteractionPurchase, "prod_123", "var_1")
	if key != "evt_123#purchase#prod_123#var_1" {
		t.Fatalf("dedupe key = %s", key)
	}
	if NewInteractionID(key) != NewInteractionID(key) {
		t.Fatal("interaction id should be deterministic")
	}
}

func validInteraction() UserInteraction {
	dedupeKey := NewInteractionDedupeKey("evt_123", InteractionProductView, "prod_123", "var_1")
	return UserInteraction{
		ID:                  NewInteractionID(dedupeKey),
		DedupeKey:           dedupeKey,
		EventID:             "evt_123",
		SourceEventType:     EventProductViewed,
		NormalizedEventType: InteractionProductView,
		Version:             1,
		Producer:            "session-service",
		UserID:              "user_123",
		ProductID:           "prod_123",
		VariantID:           "var_1",
		Weight:              1,
		Quantity:            1,
		OccurredAt:          time.Date(2026, 5, 27, 10, 30, 0, 0, time.UTC),
		ReceivedAt:          time.Date(2026, 5, 27, 10, 30, 1, 0, time.UTC),
	}
}
