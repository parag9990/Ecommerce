package domain

import (
	"errors"
	"reflect"
	"testing"
)

func TestParseOrderStatus(t *testing.T) {
	statuses := []OrderStatus{
		OrderStatusCreated,
		OrderStatusPendingPayment,
		OrderStatusPaid,
		OrderStatusPaymentFailed,
		OrderStatusPacked,
		OrderStatusShipped,
		OrderStatusDelivered,
		OrderStatusCancelled,
		OrderStatusRefunded,
	}

	for _, want := range statuses {
		got, err := ParseOrderStatus(want.String())
		if err != nil {
			t.Fatalf("ParseOrderStatus(%q) returned error: %v", want, err)
		}
		if got != want {
			t.Fatalf("ParseOrderStatus(%q) = %q, want %q", want, got, want)
		}
	}

	_, err := ParseOrderStatus("unknown")
	if !errors.Is(err, ErrUnknownOrderStatus) {
		t.Fatalf("ParseOrderStatus(unknown) error = %v, want ErrUnknownOrderStatus", err)
	}
}

func TestAllowedOrderStatusTransitions(t *testing.T) {
	statuses := []OrderStatus{
		OrderStatusCreated,
		OrderStatusPendingPayment,
		OrderStatusPaid,
		OrderStatusPaymentFailed,
		OrderStatusPacked,
		OrderStatusShipped,
		OrderStatusDelivered,
		OrderStatusCancelled,
		OrderStatusRefunded,
	}
	expected := map[OrderStatus][]OrderStatus{
		OrderStatusCreated:        {OrderStatusPendingPayment, OrderStatusPaymentFailed, OrderStatusCancelled},
		OrderStatusPendingPayment: {OrderStatusPaid, OrderStatusPaymentFailed, OrderStatusCancelled},
		OrderStatusPaid:           {OrderStatusPacked, OrderStatusCancelled, OrderStatusRefunded},
		OrderStatusPaymentFailed:  {},
		OrderStatusPacked:         {OrderStatusShipped, OrderStatusCancelled},
		OrderStatusShipped:        {OrderStatusDelivered},
		OrderStatusDelivered:      {OrderStatusRefunded},
		OrderStatusCancelled:      {OrderStatusRefunded},
		OrderStatusRefunded:       {},
	}

	for _, from := range statuses {
		got := AllowedOrderStatusTransitions(from)
		if !reflect.DeepEqual(got, expected[from]) {
			t.Fatalf("AllowedOrderStatusTransitions(%q) = %v, want %v", from, got, expected[from])
		}

		for _, to := range statuses {
			want := containsStatus(expected[from], to)
			if got := CanTransitionOrderStatus(from, to); got != want {
				t.Errorf("CanTransitionOrderStatus(%q, %q) = %t, want %t", from, to, got, want)
			}
		}
	}

	if got := AllowedOrderStatusTransitions(OrderStatus("unknown")); got != nil {
		t.Fatalf("AllowedOrderStatusTransitions(unknown) = %v, want nil", got)
	}
}

func TestAllowedOrderStatusTransitionsDoesNotExposeMutableDefinition(t *testing.T) {
	got := AllowedOrderStatusTransitions(OrderStatusCreated)
	got[0] = OrderStatusRefunded

	if CanTransitionOrderStatus(OrderStatusCreated, OrderStatusRefunded) {
		t.Fatal("mutating a returned transition slice changed the lifecycle definition")
	}
}

func TestValidateOrderStatusTransition(t *testing.T) {
	tests := []struct {
		name string
		from OrderStatus
		to   OrderStatus
		want error
	}{
		{
			name: "allowed transition",
			from: OrderStatusPendingPayment,
			to:   OrderStatusPaid,
		},
		{
			name: "invalid jump",
			from: OrderStatusPaid,
			to:   OrderStatusDelivered,
			want: ErrInvalidOrderStatusTransition,
		},
		{
			name: "unchanged status",
			from: OrderStatusPaid,
			to:   OrderStatusPaid,
			want: ErrOrderStatusUnchanged,
		},
		{
			name: "payment creation failure",
			from: OrderStatusCreated,
			to:   OrderStatusPaymentFailed,
		},
		{
			name: "payment result failure",
			from: OrderStatusPendingPayment,
			to:   OrderStatusPaymentFailed,
		},
		{
			name: "unknown current status",
			from: OrderStatus("unknown"),
			to:   OrderStatusCancelled,
			want: ErrUnknownOrderStatus,
		},
		{
			name: "unknown next status",
			from: OrderStatusCreated,
			to:   OrderStatus("unknown"),
			want: ErrUnknownOrderStatus,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateOrderStatusTransition(test.from, test.to)
			if !errors.Is(err, test.want) {
				t.Fatalf("ValidateOrderStatusTransition(%q, %q) error = %v, want %v", test.from, test.to, err, test.want)
			}
		})
	}
}

func TestParseOrderStatusActorType(t *testing.T) {
	actorTypes := []OrderStatusActorType{
		OrderStatusActorSystem,
		OrderStatusActorBuyer,
		OrderStatusActorSeller,
		OrderStatusActorAdmin,
		OrderStatusActorPaymentService,
		OrderStatusActorLogistics,
	}

	for _, want := range actorTypes {
		got, err := ParseOrderStatusActorType(want.String())
		if err != nil {
			t.Fatalf("ParseOrderStatusActorType(%q) returned error: %v", want, err)
		}
		if got != want {
			t.Fatalf("ParseOrderStatusActorType(%q) = %q, want %q", want, got, want)
		}
	}

	_, err := ParseOrderStatusActorType("warehouse")
	if !errors.Is(err, ErrUnknownOrderStatusActorType) {
		t.Fatalf("ParseOrderStatusActorType(warehouse) error = %v, want ErrUnknownOrderStatusActorType", err)
	}
}

func containsStatus(statuses []OrderStatus, want OrderStatus) bool {
	for _, status := range statuses {
		if status == want {
			return true
		}
	}

	return false
}
