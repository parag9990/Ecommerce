package repository

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/example/ecommerce-platform/backend/services/order-service/internal/domain"
)

func TestMySQLOrderRepositoryCreatesOrderAndAttachesClaimInOneTransaction(t *testing.T) {
	db, mock := newMockDB(t)
	repository, _ := NewMySQLOrderRepository(db)
	order, claim := validOrderAndClaim()
	event := validCreatedOutboxEvent(order)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO orders").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO order_items").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO order_status_history").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO order_outbox_events").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE order_idempotency_keys").
		WithArgs(order.OrderID, claim.UserID, claim.Key, claim.RequestHash).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := repository.CreateOrderAndAttachClaim(context.Background(), order, claim, event); err != nil {
		t.Fatalf("CreateOrderAndAttachClaim() error = %v", err)
	}
	assertMockExpectations(t, mock)
}

func TestMySQLOrderRepositoryRollsBackWhenClaimCannotBeAttached(t *testing.T) {
	db, mock := newMockDB(t)
	repository, _ := NewMySQLOrderRepository(db)
	order, claim := validOrderAndClaim()
	event := validCreatedOutboxEvent(order)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO orders").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO order_items").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO order_status_history").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO order_outbox_events").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE order_idempotency_keys").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	err := repository.CreateOrderAndAttachClaim(context.Background(), order, claim, event)
	if !errors.Is(err, domain.ErrCheckoutInProgress) {
		t.Fatalf("CreateOrderAndAttachClaim() error = %v, want ErrCheckoutInProgress", err)
	}
	assertMockExpectations(t, mock)
}

func TestMySQLOrderRepositoryReportsUnknownCommitOutcome(t *testing.T) {
	db, mock := newMockDB(t)
	repository, _ := NewMySQLOrderRepository(db)
	order, claim := validOrderAndClaim()
	event := validCreatedOutboxEvent(order)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO orders").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO order_items").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO order_status_history").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO order_outbox_events").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE order_idempotency_keys").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit().WillReturnError(errors.New("commit response lost"))

	err := repository.CreateOrderAndAttachClaim(context.Background(), order, claim, event)
	if !errors.Is(err, domain.ErrOrderCreateOutcomeUnknown) {
		t.Fatalf("CreateOrderAndAttachClaim() error = %v, want ErrOrderCreateOutcomeUnknown", err)
	}
	assertMockExpectations(t, mock)
}

func TestMySQLOrderRepositoryCancelsOrderWithOutboxEventInOneTransaction(t *testing.T) {
	db, mock := newMockDB(t)
	repository, _ := NewMySQLOrderRepository(db)
	now := time.Date(2026, time.May, 26, 10, 0, 0, 0, time.UTC)
	from := domain.OrderStatusPendingPayment
	history := domain.OrderStatusHistoryEntry{
		ID:         "osh_cancelled",
		OrderID:    "ord_1",
		FromStatus: &from,
		ToStatus:   domain.OrderStatusCancelled,
		Reason:     "buyer_requested",
		ActorType:  domain.OrderStatusActorBuyer,
		ActorID:    "user_1",
		CreatedAt:  now,
	}
	event := domain.OutboxEvent{
		EventID: "evt_cancelled", EventType: "OrderCancelled", Version: 1,
		AggregateType: "order", AggregateID: "ord_1", RoutingKey: "order.cancelled",
		DeduplicationKey: "osh_cancelled:OrderCancelled", SourceHistoryID: "osh_cancelled",
		Payload: []byte(`{"event_id":"evt_cancelled","event_type":"OrderCancelled"}`),
		Status:  domain.OutboxStatusPending, CreatedAt: now, UpdatedAt: now,
	}
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE orders").
		WithArgs("cancelled", "ord_1", "pending_payment").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE order_items").
		WithArgs("cancelled", "ord_1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO order_status_history").
		WithArgs("osh_cancelled", "ord_1", "pending_payment", "cancelled", "buyer_requested", "buyer", "user_1", nil, now).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO order_outbox_events").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	expectGetCancelledOrder(mock, now)

	order, err := repository.CancelOrder(context.Background(), "ord_1", from, history, event)
	if err != nil {
		t.Fatalf("CancelOrder() error = %v", err)
	}
	if order.Status != domain.OrderStatusCancelled || len(order.Items) != 1 ||
		order.Items[0].FulfillmentStatus != domain.FulfillmentStatusCancelled {
		t.Fatalf("order = %+v, want cancelled order with cancelled item", order)
	}
	assertMockExpectations(t, mock)
}

func validCreatedOutboxEvent(order domain.Order) domain.OutboxEvent {
	return domain.OutboxEvent{
		EventID:          "evt_created",
		EventType:        "OrderCreated",
		Version:          1,
		AggregateType:    "order",
		AggregateID:      order.OrderID,
		RoutingKey:       "order.created",
		DeduplicationKey: order.OrderID + ":created",
		SourceHistoryID:  order.InitialHistory.ID,
		Payload:          []byte(`{"event_id":"evt_created","event_type":"OrderCreated"}`),
		Status:           domain.OutboxStatusPending,
		CreatedAt:        order.InitialHistory.CreatedAt,
		UpdatedAt:        order.InitialHistory.CreatedAt,
	}
}

func validOrderAndClaim() (domain.Order, domain.IdempotencyClaim) {
	now := time.Date(2026, time.May, 26, 10, 0, 0, 0, time.UTC)
	order := domain.Order{
		OrderID: "ord_1", UserID: "user_1", CartID: "cart_1", Status: domain.OrderStatusCreated,
		Currency: "INR", SubtotalAmount: 1000, TotalAmount: 1000,
		AddressSnapshot: domain.AddressSnapshot{
			RecipientName: "Buyer", Line1: "101 Main Street", City: "Delhi", PostalCode: "110001", Country: "IN",
		},
		InventoryReservationID: "res_1", InventoryReservedUntil: now.Add(10 * time.Minute),
		Items: []domain.OrderItem{{
			OrderItemID: "oi_1", OrderID: "ord_1", SellerID: "seller_1", ProductID: "product_1",
			VariantID: "variant_1", SKU: "SKU-1", TitleSnapshot: "Product", Quantity: 1,
			Currency: "INR", UnitAmount: 1000, TotalAmount: 1000, FulfillmentStatus: domain.FulfillmentStatusPending,
		}},
		InitialHistory: domain.OrderStatusHistoryEntry{
			ID: "osh_1", OrderID: "ord_1", ToStatus: domain.OrderStatusCreated, Reason: "cart_to_order_created",
			ActorType: domain.OrderStatusActorBuyer, ActorID: "user_1", CreatedAt: now,
		},
	}
	claim := domain.IdempotencyClaim{
		UserID: "user_1", Key: "key_1", RequestHash: strings.Repeat("a", 64),
		Status: domain.IdempotencyStatusProcessing, Decision: domain.ClaimAcquired, ExpiresAt: now.Add(24 * time.Hour),
	}
	return order, claim
}

func expectGetCancelledOrder(mock sqlmock.Sqlmock, now time.Time) {
	mock.ExpectQuery("FROM orders").
		WithArgs("ord_1").
		WillReturnRows(sqlmock.NewRows([]string{
			"order_id", "user_id", "cart_id", "status", "currency", "subtotal_amount",
			"discount_amount", "shipping_amount", "tax_amount", "total_amount",
			"address_snapshot", "coupon_code", "payment_id", "inventory_reservation_id",
			"inventory_reserved_until", "created_at", "updated_at",
		}).AddRow(
			"ord_1", "user_1", "cart_1", "cancelled", "INR", int64(1000),
			int64(0), int64(0), int64(0), int64(1000),
			[]byte(`{"recipient_name":"Buyer","line1":"Street","city":"Delhi","postal_code":"110001","country":"IN"}`),
			nil, "pay_1", "res_1", now.Add(time.Minute), now, now,
		))
	mock.ExpectQuery("FROM order_items").
		WithArgs("ord_1").
		WillReturnRows(sqlmock.NewRows([]string{
			"order_item_id", "order_id", "seller_id", "product_id", "variant_id",
			"sku", "title_snapshot", "image_url_snapshot", "quantity", "currency",
			"unit_amount", "discount_amount", "tax_amount", "total_amount", "fulfillment_status",
		}).AddRow(
			"oi_1", "ord_1", "seller_1", "product_1", "variant_1",
			"SKU-1", "Product", nil, int32(1), "INR",
			int64(1000), int64(0), int64(0), int64(1000), "cancelled",
		))
}
