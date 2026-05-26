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
