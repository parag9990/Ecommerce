package events

import (
	"testing"

	"product-service/internal/domain"
)

func TestProductRoutingKeysMatchSearchConsumerBindings(t *testing.T) {
	tests := map[domain.ProductEventType]string{
		domain.ProductEventCreated:          "product.created",
		domain.ProductEventUpdated:          "product.updated",
		domain.ProductEventPublished:        "product.published",
		domain.ProductEventUnpublished:      "product.unpublished",
		domain.ProductEventInventoryChanged: "product.inventory_changed",
	}
	for eventType, expected := range tests {
		if actual := productRoutingKey(string(eventType)); actual != expected {
			t.Errorf("routing key for %s = %q, want %q", eventType, actual, expected)
		}
	}
}
