package usecase

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/order-service/internal/domain"
)

func TestInitiateOrderPaymentAttachesIntentUsingPersistedAmount(t *testing.T) {
	now := paymentTestTime()
	orders := &fakePaymentRepository{order: payableOrder(now)}
	payments := &fakePaymentClient{intent: validPaymentIntent(now)}
	inventory := &fakePaymentInventory{}
	usecase := newInitiatePaymentUsecase(t, orders, payments, inventory, now)

	result, err := usecase.Execute(context.Background(), InitiateOrderPaymentCommand{OrderID: "ord_1", UserID: "user_1"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.OrderStatus != domain.OrderStatusPendingPayment || result.PaymentID != "pay_1" ||
		orders.order.Status != domain.OrderStatusPendingPayment || orders.order.PaymentID != "pay_1" {
		t.Fatalf("payment result/order = %+v/%+v, want attached pending payment", result, orders.order)
	}
	if payments.request.Amount != 299900 || payments.request.Currency != "INR" ||
		payments.request.IdempotencyKey != "payment_intent:ord_1:1" ||
		payments.request.ReturnURL != "https://shop.example.test/payment/return" {
		t.Fatalf("payment request = %+v, want persisted amount and stable contract values", payments.request)
	}
	if orders.histories[0].Reason != "payment_intent_created" || inventory.releaseCalls != 0 {
		t.Fatalf("histories/releases = %+v/%d, want intent history without release", orders.histories, inventory.releaseCalls)
	}
}

func TestInitiateOrderPaymentReconstructsExistingIntentForCheckoutResume(t *testing.T) {
	now := paymentTestTime()
	order := payableOrder(now)
	order.Status = domain.OrderStatusPendingPayment
	order.PaymentID = "pay_1"
	orders := &fakePaymentRepository{order: order}
	payments := &fakePaymentClient{intent: validPaymentIntent(now)}
	usecase := newInitiatePaymentUsecase(t, orders, payments, &fakePaymentInventory{}, now)

	result, err := usecase.Execute(context.Background(), InitiateOrderPaymentCommand{OrderID: "ord_1", UserID: "user_1"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.PaymentID != "pay_1" || payments.calls != 1 ||
		payments.request.IdempotencyKey != "payment_intent:ord_1:1" ||
		orders.order.Status != domain.OrderStatusPendingPayment {
		t.Fatalf("result/payment/order = %+v/%+v/%+v, want deterministic existing intent reconstruction", result, payments.request, orders.order)
	}
}

func TestInitiateOrderPaymentRejectsUnpayableOrderBeforeCreatingIntent(t *testing.T) {
	now := paymentTestTime()
	tests := []struct {
		name   string
		mutate func(*domain.Order)
		userID string
		want   error
	}{
		{
			name:   "wrong buyer",
			userID: "another_user",
			want:   domain.ErrOrderForbidden,
		},
		{
			name: "expired reservation",
			mutate: func(order *domain.Order) {
				order.InventoryReservedUntil = now
			},
			userID: "user_1",
			want:   domain.ErrInventoryReservationExpired,
		},
		{
			name: "unsupported currency",
			mutate: func(order *domain.Order) {
				order.Currency = "EUR"
			},
			userID: "user_1",
			want:   domain.ErrUnsupportedCurrency,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			order := payableOrder(now)
			if test.mutate != nil {
				test.mutate(&order)
			}
			orders := &fakePaymentRepository{order: order}
			payments := &fakePaymentClient{intent: validPaymentIntent(now)}
			usecase := newInitiatePaymentUsecase(t, orders, payments, &fakePaymentInventory{}, now)

			_, err := usecase.Execute(context.Background(), InitiateOrderPaymentCommand{OrderID: "ord_1", UserID: test.userID})
			if !errors.Is(err, test.want) {
				t.Fatalf("Execute() error = %v, want %v", err, test.want)
			}
			if payments.calls != 0 || orders.attachCalls != 0 {
				t.Fatal("unpayable order reached Payment Service or status transition")
			}
		})
	}
}

func TestInitiateOrderPaymentTerminalFailureMarksFailedAndReleasesInventory(t *testing.T) {
	now := paymentTestTime()
	orders := &fakePaymentRepository{order: payableOrder(now)}
	payments := &fakePaymentClient{err: errors.New("provider rejected request")}
	inventory := &fakePaymentInventory{}
	usecase := newInitiatePaymentUsecase(t, orders, payments, inventory, now)

	_, err := usecase.Execute(context.Background(), InitiateOrderPaymentCommand{OrderID: "ord_1", UserID: "user_1"})
	if !errors.Is(err, domain.ErrPaymentIntentCreationFailed) {
		t.Fatalf("Execute() error = %v, want ErrPaymentIntentCreationFailed", err)
	}
	if orders.order.Status != domain.OrderStatusPaymentFailed || inventory.releaseCalls != 1 ||
		inventory.releaseRequest.IdempotencyKey != "inventory_release:ord_1:payment_intent_failed" {
		t.Fatalf("order/release = %+v/%+v, want compensated terminal intent failure", orders.order, inventory.releaseRequest)
	}
}

func TestInitiateOrderPaymentUnknownOutcomeKeepsReservationPendingResolution(t *testing.T) {
	now := paymentTestTime()
	orders := &fakePaymentRepository{order: payableOrder(now)}
	payments := &fakePaymentClient{err: domain.ErrPaymentIntentPendingResolution}
	inventory := &fakePaymentInventory{}
	usecase := newInitiatePaymentUsecase(t, orders, payments, inventory, now)

	_, err := usecase.Execute(context.Background(), InitiateOrderPaymentCommand{OrderID: "ord_1", UserID: "user_1"})
	if !errors.Is(err, domain.ErrPaymentIntentPendingResolution) {
		t.Fatalf("Execute() error = %v, want ErrPaymentIntentPendingResolution", err)
	}
	if orders.order.Status != domain.OrderStatusCreated || orders.failureCalls != 0 || inventory.releaseCalls != 0 {
		t.Fatalf("unknown intent outcome finalized order/inventory: order=%+v failures=%d releases=%d", orders.order, orders.failureCalls, inventory.releaseCalls)
	}
}

func TestApplyPaymentResultCapturedMarksPaidAndCommitsInventory(t *testing.T) {
	now := paymentTestTime()
	order := payableOrder(now)
	order.Status = domain.OrderStatusPendingPayment
	order.PaymentID = "pay_1"
	orders := &fakePaymentRepository{order: order}
	inventory := &fakePaymentInventory{}
	usecase := newApplyPaymentUsecase(t, orders, inventory, now)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := usecase.Execute(ctx, validPaymentResultCommand("captured", now))
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if orders.order.Status != domain.OrderStatusPaid || inventory.commitCalls != 1 ||
		inventory.commitRequest.IdempotencyKey != "inventory_commit:ord_1" ||
		inventory.actionContextErr != nil {
		t.Fatalf("order/commit = %+v/%+v contextErr=%v, want paid and committed inventory", orders.order, inventory.commitRequest, inventory.actionContextErr)
	}
	if orders.histories[0].Metadata["provider_event_id"] != "evt_1" {
		t.Fatalf("history = %+v, want payment event audit metadata", orders.histories[0])
	}
	if orders.events[0].EventType != "OrderPaid" || orders.events[0].AggregateID != "ord_1" {
		t.Fatalf("outbox events = %+v, want OrderPaid for order", orders.events)
	}
}

func TestApplyPaymentResultFailedMarksPaymentFailedAndReleasesInventory(t *testing.T) {
	now := paymentTestTime()
	order := payableOrder(now)
	order.Status = domain.OrderStatusPendingPayment
	order.PaymentID = "pay_1"
	orders := &fakePaymentRepository{order: order}
	inventory := &fakePaymentInventory{}
	usecase := newApplyPaymentUsecase(t, orders, inventory, now)

	err := usecase.Execute(context.Background(), validPaymentResultCommand("failed", now))
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if orders.order.Status != domain.OrderStatusPaymentFailed || inventory.releaseCalls != 1 ||
		inventory.releaseRequest.Reason != "provider_payment_failed" {
		t.Fatalf("order/release = %+v/%+v, want payment failure inventory release", orders.order, inventory.releaseRequest)
	}
}

func TestApplyPaymentResultRetriesInventoryForDuplicateCapturedResult(t *testing.T) {
	now := paymentTestTime()
	for _, status := range []domain.OrderStatus{
		domain.OrderStatusPaid,
		domain.OrderStatusPacked,
		domain.OrderStatusShipped,
		domain.OrderStatusDelivered,
		domain.OrderStatusRefunded,
	} {
		t.Run(status.String(), func(t *testing.T) {
			orders := &fakePaymentRepository{order: withPaymentState(payableOrder(now), status)}
			inventory := &fakePaymentInventory{}
			usecase := newApplyPaymentUsecase(t, orders, inventory, now)

			if err := usecase.Execute(context.Background(), validPaymentResultCommand("captured", now)); err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if orders.transitionCalls != 0 || inventory.commitCalls != 1 || inventory.releaseCalls != 0 {
				t.Fatalf("transitions/commits/releases = %d/%d/%d, want 0/1/0", orders.transitionCalls, inventory.commitCalls, inventory.releaseCalls)
			}
		})
	}
}

func TestApplyPaymentResultRetriesInventoryForDuplicateFailedResult(t *testing.T) {
	now := paymentTestTime()
	orders := &fakePaymentRepository{order: withPaymentState(payableOrder(now), domain.OrderStatusPaymentFailed)}
	inventory := &fakePaymentInventory{}
	usecase := newApplyPaymentUsecase(t, orders, inventory, now)

	if err := usecase.Execute(context.Background(), validPaymentResultCommand("failed", now)); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if orders.transitionCalls != 0 || inventory.commitCalls != 0 || inventory.releaseCalls != 1 {
		t.Fatalf("transitions/commits/releases = %d/%d/%d, want 0/0/1", orders.transitionCalls, inventory.commitCalls, inventory.releaseCalls)
	}
}

func TestApplyPaymentResultRejectsMismatchedPaymentWithoutInventoryAction(t *testing.T) {
	now := paymentTestTime()
	orders := &fakePaymentRepository{order: withPaymentState(payableOrder(now), domain.OrderStatusPendingPayment)}
	inventory := &fakePaymentInventory{}
	usecase := newApplyPaymentUsecase(t, orders, inventory, now)
	command := validPaymentResultCommand("captured", now)
	command.Amount++

	err := usecase.Execute(context.Background(), command)
	if !errors.Is(err, domain.ErrPaymentAmountMismatch) {
		t.Fatalf("Execute() error = %v, want %v", err, domain.ErrPaymentAmountMismatch)
	}
	if orders.transitionCalls != 0 || inventory.commitCalls != 0 || inventory.releaseCalls != 0 {
		t.Fatal("mismatched payment caused a status transition or inventory action")
	}
}

type fakePaymentRepository struct {
	order           domain.Order
	histories       []domain.OrderStatusHistoryEntry
	events          []domain.OutboxEvent
	attachCalls     int
	failureCalls    int
	transitionCalls int
}

func (f *fakePaymentRepository) GetForPayment(context.Context, string) (domain.Order, error) {
	return f.order, nil
}

func (f *fakePaymentRepository) AttachPaymentIntent(_ context.Context, _ string, paymentID string, history domain.OrderStatusHistoryEntry) (bool, error) {
	f.attachCalls++
	if f.order.Status != domain.OrderStatusCreated || f.order.PaymentID != "" {
		return false, nil
	}
	f.order.Status = domain.OrderStatusPendingPayment
	f.order.PaymentID = paymentID
	f.histories = append(f.histories, history)
	return true, nil
}

func (f *fakePaymentRepository) MarkPaymentFailedFromCreated(_ context.Context, _ string, history domain.OrderStatusHistoryEntry) (bool, error) {
	f.failureCalls++
	if f.order.Status != domain.OrderStatusCreated || f.order.PaymentID != "" {
		return false, nil
	}
	f.order.Status = domain.OrderStatusPaymentFailed
	f.histories = append(f.histories, history)
	return true, nil
}

func (f *fakePaymentRepository) TransitionPaymentResult(_ context.Context, _ string, _ string, from domain.OrderStatus, to domain.OrderStatus, history domain.OrderStatusHistoryEntry, event *domain.OutboxEvent) (bool, error) {
	f.transitionCalls++
	if f.order.Status != from {
		return false, nil
	}
	f.order.Status = to
	f.histories = append(f.histories, history)
	if event != nil {
		f.events = append(f.events, *event)
	}
	return true, nil
}

type fakePaymentClient struct {
	intent  *CreatePaymentIntentResponse
	err     error
	request CreatePaymentIntentRequest
	calls   int
}

func (f *fakePaymentClient) CreatePaymentIntent(_ context.Context, request CreatePaymentIntentRequest) (*CreatePaymentIntentResponse, error) {
	f.calls++
	f.request = request
	return f.intent, f.err
}

type fakePaymentInventory struct {
	commitCalls      int
	releaseCalls     int
	commitRequest    CommitInventoryRequest
	releaseRequest   ReleaseInventoryRequest
	actionContextErr error
}

func (f *fakePaymentInventory) CommitInventory(ctx context.Context, request CommitInventoryRequest) error {
	f.commitCalls++
	f.commitRequest = request
	f.actionContextErr = ctx.Err()
	return nil
}

func (f *fakePaymentInventory) ReleaseInventory(ctx context.Context, request ReleaseInventoryRequest) error {
	f.releaseCalls++
	f.releaseRequest = request
	f.actionContextErr = ctx.Err()
	return nil
}

func newInitiatePaymentUsecase(t *testing.T, orders PaymentOrderRepository, payments PaymentClient, inventory InventoryReleaser, now time.Time) *InitiateOrderPaymentUsecase {
	t.Helper()
	usecase, err := NewInitiateOrderPaymentUsecase(orders, payments, inventory, &sequenceIDGenerator{values: []string{"osh_payment"}}, InitiateOrderPaymentConfig{
		ReturnURL:               "https://shop.example.test/payment/return",
		AllowedCurrencies:       []string{"INR", "USD"},
		InventoryReleaseTimeout: time.Second,
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("NewInitiateOrderPaymentUsecase() error = %v", err)
	}
	usecase.WithClock(fixedClock{now: now})
	return usecase
}

func newApplyPaymentUsecase(t *testing.T, orders PaymentOrderRepository, inventory InventoryClient, now time.Time) *ApplyPaymentResultUsecase {
	t.Helper()
	usecase, err := NewApplyPaymentResultUsecase(orders, inventory, &sequenceIDGenerator{values: []string{"osh_result", "evt_paid"}}, ApplyPaymentResultConfig{
		InventoryActionTimeout: time.Second,
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("NewApplyPaymentResultUsecase() error = %v", err)
	}
	usecase.WithClock(fixedClock{now: now})
	return usecase
}

func payableOrder(now time.Time) domain.Order {
	return domain.Order{
		OrderID:                "ord_1",
		UserID:                 "user_1",
		Status:                 domain.OrderStatusCreated,
		Currency:               "INR",
		TotalAmount:            299900,
		InventoryReservationID: "res_1",
		InventoryReservedUntil: now.Add(10 * time.Minute),
	}
}

func withPaymentState(order domain.Order, status domain.OrderStatus) domain.Order {
	order.Status = status
	order.PaymentID = "pay_1"
	return order
}

func validPaymentIntent(now time.Time) *CreatePaymentIntentResponse {
	return &CreatePaymentIntentResponse{
		PaymentID:         "pay_1",
		Status:            "requires_action",
		Provider:          "provider",
		ClientActionToken: "safe_action_token",
		ExpiresAt:         now.Add(5 * time.Minute),
	}
}

func validPaymentResultCommand(result string, now time.Time) ApplyPaymentResultCommand {
	return ApplyPaymentResultCommand{
		PaymentID:       "pay_1",
		OrderID:         "ord_1",
		Result:          result,
		Amount:          299900,
		Currency:        "INR",
		ProviderEventID: "evt_1",
		OccurredAt:      now,
	}
}

func paymentTestTime() time.Time {
	return time.Date(2026, time.May, 26, 10, 0, 0, 0, time.UTC)
}
