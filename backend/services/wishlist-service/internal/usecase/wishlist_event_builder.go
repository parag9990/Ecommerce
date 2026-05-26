package usecase

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"ecommerce/backend/services/wishlist-service/internal/domain"
)

const DefaultRecommendationEventsTopic = "recommendation.events"

func newWishlistAnalyticsEvent(
	eventID string,
	topic string,
	eventType domain.WishlistEventType,
	userID string,
	item domain.WishlistItem,
	traceID string,
	occurredAt time.Time,
) domain.WishlistAnalyticsEvent {
	action := domain.WishlistEventActionAdd
	if eventType == domain.WishlistEventItemRemoved {
		action = domain.WishlistEventActionRemove
	}
	if strings.TrimSpace(topic) == "" {
		topic = DefaultRecommendationEventsTopic
	}

	return domain.WishlistAnalyticsEvent{
		EventID:   eventID,
		EventType: eventType,
		Version:   domain.WishlistEventVersion,
		Topic:     topic,
		Payload: domain.WishlistAnalyticsPayload{
			UserID:         userID,
			ProductID:      item.ProductID,
			VariantID:      item.VariantID,
			Action:         action,
			Source:         domain.WishlistEventSource,
			Availability:   item.Availability.Normalized(),
			LastKnownPrice: copyMoney(item.LastKnownPrice),
		},
		TraceID:     normalizeID(traceID),
		Status:      domain.WishlistEventPending,
		Attempts:    0,
		NextRetryAt: occurredAt.UTC(),
		OccurredAt:  occurredAt.UTC(),
		CreatedAt:   occurredAt.UTC(),
		UpdatedAt:   occurredAt.UTC(),
	}
}

func defaultWishlistEventID() string {
	var randomBytes [16]byte
	if _, err := rand.Read(randomBytes[:]); err == nil {
		return "evt_wish_" + hex.EncodeToString(randomBytes[:])
	}
	return fmt.Sprintf("evt_wish_%d", time.Now().UTC().UnixNano())
}

func copyMoney(money *domain.Money) *domain.Money {
	if money == nil {
		return nil
	}
	copied := *money
	copied.Currency = strings.ToUpper(strings.TrimSpace(copied.Currency))
	return &copied
}
