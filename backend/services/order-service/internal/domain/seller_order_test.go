package domain

import (
	"errors"
	"testing"
)

func TestDeriveSellerFulfillmentStatus(t *testing.T) {
	tests := []struct {
		name  string
		items []SellerOrderItemView
		want  SellerFulfillmentStatus
	}{
		{
			name: "pending",
			items: []SellerOrderItemView{
				{FulfillmentStatus: FulfillmentStatusPending},
				{FulfillmentStatus: FulfillmentStatusPending},
			},
			want: SellerFulfillmentStatusPending,
		},
		{
			name: "partially shipped",
			items: []SellerOrderItemView{
				{FulfillmentStatus: FulfillmentStatusPacked},
				{FulfillmentStatus: FulfillmentStatusShipped},
			},
			want: SellerFulfillmentStatusPartiallyShipped,
		},
		{
			name: "delivered ignores returned inactive line",
			items: []SellerOrderItemView{
				{FulfillmentStatus: FulfillmentStatusDelivered},
				{FulfillmentStatus: FulfillmentStatusReturned},
			},
			want: SellerFulfillmentStatusDelivered,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := DeriveSellerFulfillmentStatus(test.items); got != test.want {
				t.Fatalf("DeriveSellerFulfillmentStatus() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestAggregateParentFulfillmentStatusWaitsForAllSellers(t *testing.T) {
	got, changed := AggregateParentFulfillmentStatus(OrderStatusPaid, []FulfillmentStatus{
		FulfillmentStatusPacked,
		FulfillmentStatusPending,
	})
	if got != OrderStatusPaid || changed {
		t.Fatalf("AggregateParentFulfillmentStatus(partial packed) = %q/%t, want paid/false", got, changed)
	}
	got, changed = AggregateParentFulfillmentStatus(OrderStatusPaid, []FulfillmentStatus{
		FulfillmentStatusPacked,
		FulfillmentStatusPacked,
	})
	if got != OrderStatusPacked || !changed {
		t.Fatalf("AggregateParentFulfillmentStatus(all packed) = %q/%t, want packed/true", got, changed)
	}
}

func TestValidateSellerItemTransitionRejectsSkippedSteps(t *testing.T) {
	if err := ValidateSellerItemTransition(FulfillmentStatusPending, FulfillmentStatusShipped); !errors.Is(err, ErrInvalidFulfillmentTransition) {
		t.Fatalf("ValidateSellerItemTransition(skip) error = %v, want ErrInvalidFulfillmentTransition", err)
	}
	if err := ValidateSellerItemTransition(FulfillmentStatusPacked, FulfillmentStatusShipped); err != nil {
		t.Fatalf("ValidateSellerItemTransition(next) error = %v", err)
	}
}
