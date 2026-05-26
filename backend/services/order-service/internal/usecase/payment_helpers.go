package usecase

import (
	"fmt"
	"strings"

	"github.com/example/ecommerce-platform/backend/services/order-service/internal/domain"
)

func newPaymentHistory(
	ids IDGenerator,
	clock Clock,
	orderID string,
	from domain.OrderStatus,
	to domain.OrderStatus,
	reason string,
	actor domain.OrderStatusActorType,
	actorID string,
	metadata map[string]any,
) (domain.OrderStatusHistoryEntry, error) {
	historyID, err := ids.NewID("osh")
	if err != nil {
		return domain.OrderStatusHistoryEntry{}, fmt.Errorf("generate order history id: %w", err)
	}
	fromStatus := from
	return domain.OrderStatusHistoryEntry{
		ID:         historyID,
		OrderID:    strings.TrimSpace(orderID),
		FromStatus: &fromStatus,
		ToStatus:   to,
		Reason:     reason,
		ActorType:  actor,
		ActorID:    actorID,
		Metadata:   metadata,
		CreatedAt:  clock.Now().UTC(),
	}, nil
}

func normalizeCurrencyCode(value string) (string, bool) {
	value = strings.ToUpper(strings.TrimSpace(value))
	if len(value) != 3 {
		return "", false
	}
	for _, character := range value {
		if character < 'A' || character > 'Z' {
			return "", false
		}
	}
	return value, true
}
