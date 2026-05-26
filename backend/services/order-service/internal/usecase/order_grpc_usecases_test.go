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

func TestCreateOrderUsecaseComposesCheckoutAndPayment(t *testing.T) {
	created := &fakeCreateFromCart{result: &CreateOrderFromCartResult{OrderID: "ord_1"}}
	payment := &fakePaymentInitiation{result: &InitiateOrderPaymentResult{
		PaymentID: "pay_1", ClientActionToken: "action", ExpiresAt: paymentTestTime().Add(time.Minute),
	}}
	reader := &fakeOrderReader{order: domain.Order{OrderID: "ord_1", UserID: "user_1", Status: domain.OrderStatusPendingPayment}}
	idempotency := &fakeIdempotencyRepository{decision: domain.ClaimAcquired}
	workflow := newCreateWorkflow(t, created, payment, reader, idempotency)
	result, err := workflow.Execute(context.Background(), CreateOrderCommand{
		UserID: "user_1", CartID: "cart_1", IdempotencyKey: "request_1",
		ShippingAddress: validCommand().ShippingAddress,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if created.command.UserID != "user_1" || payment.command.OrderID != "ord_1" ||
		result.Order.Status != domain.OrderStatusPendingPayment || result.PaymentAction.PaymentID != "pay_1" ||
		idempotency.completeCalls != 1 || created.command.Claim.Decision != domain.ClaimAcquired {
		t.Fatalf("workflow result/calls = %+v/%+v/%+v, want delegated paid-ready response", created.command, payment.command, result)
	}
}

func TestCreateOrderUsecaseReplaysCompletedOrderWithoutCreatingAnotherOrder(t *testing.T) {
	created := &fakeCreateFromCart{}
	payment := &fakePaymentInitiation{}
	reader := &fakeOrderReader{order: domain.Order{OrderID: "ord_existing", UserID: "user_1", Status: domain.OrderStatusPaid}}
	idempotency := &fakeIdempotencyRepository{decision: domain.ClaimReplay, orderID: "ord_existing"}
	workflow := newCreateWorkflow(t, created, payment, reader, idempotency)

	result, err := workflow.Execute(context.Background(), validCreateOrderCommand())
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.Order.OrderID != "ord_existing" || created.calls != 0 || payment.calls != 0 ||
		idempotency.completeCalls != 0 {
		t.Fatalf("result/calls = %+v checkout=%d payment=%d complete=%d, want pure completed replay", result, created.calls, payment.calls, idempotency.completeCalls)
	}
}

func TestCreateOrderUsecaseResumesBoundProcessingOrderWithoutCreatingAnotherOrder(t *testing.T) {
	created := &fakeCreateFromCart{}
	payment := &fakePaymentInitiation{result: &InitiateOrderPaymentResult{PaymentID: "pay_1"}}
	reader := &fakeOrderReader{order: domain.Order{OrderID: "ord_existing", UserID: "user_1", Status: domain.OrderStatusPendingPayment}}
	idempotency := &fakeIdempotencyRepository{decision: domain.ClaimResume, orderID: "ord_existing"}
	workflow := newCreateWorkflow(t, created, payment, reader, idempotency)

	_, err := workflow.Execute(context.Background(), validCreateOrderCommand())
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if created.calls != 0 || payment.calls != 1 || idempotency.completeCalls != 1 {
		t.Fatalf("checkout/payment/complete calls = %d/%d/%d, want post-order resume only", created.calls, payment.calls, idempotency.completeCalls)
	}
}

func TestCreateOrderUsecaseRejectsDuplicateInProgressAndChangedPayload(t *testing.T) {
	tests := []struct {
		name        string
		idempotency *fakeIdempotencyRepository
		want        error
	}{
		{name: "in progress", idempotency: &fakeIdempotencyRepository{decision: domain.ClaimInProgress}, want: domain.ErrCheckoutInProgress},
		{name: "changed request", idempotency: &fakeIdempotencyRepository{claimErr: domain.ErrIdempotencyConflict}, want: domain.ErrIdempotencyConflict},
		{name: "prior failed", idempotency: &fakeIdempotencyRepository{decision: domain.ClaimFailed}, want: domain.ErrCheckoutFailed},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			created := &fakeCreateFromCart{}
			payment := &fakePaymentInitiation{}
			workflow := newCreateWorkflow(t, created, payment, &fakeOrderReader{}, test.idempotency)

			_, err := workflow.Execute(context.Background(), validCreateOrderCommand())
			if !errors.Is(err, test.want) || created.calls != 0 || payment.calls != 0 {
				t.Fatalf("Execute() error/calls = %v/%d/%d, want %v without workflow", err, created.calls, payment.calls, test.want)
			}
		})
	}
}

func TestCreateOrderUsecaseMarksConfirmedPreOrderFailureTerminal(t *testing.T) {
	created := &fakeCreateFromCart{err: domain.ErrInventoryUnavailable}
	idempotency := &fakeIdempotencyRepository{decision: domain.ClaimAcquired}
	workflow := newCreateWorkflow(t, created, &fakePaymentInitiation{}, &fakeOrderReader{}, idempotency)

	_, err := workflow.Execute(context.Background(), validCreateOrderCommand())
	if !errors.Is(err, domain.ErrInventoryUnavailable) || idempotency.failedCalls != 1 {
		t.Fatalf("Execute() error/failed calls = %v/%d, want terminal claim failure", err, idempotency.failedCalls)
	}
}

func TestCheckoutRequestHashNormalizesEquivalentInputAndDetectsChanges(t *testing.T) {
	first := validCreateOrderCommand()
	second := first
	second.CartID = " cart_1 "
	second.ShippingAddress.Country = " in "
	changed := first
	changed.ShippingAddress.Line1 = "Office Street"

	firstHash, _ := checkoutRequestHash(normalizeCreateOrderCommand(first))
	secondHash, _ := checkoutRequestHash(normalizeCreateOrderCommand(second))
	changedHash, _ := checkoutRequestHash(normalizeCreateOrderCommand(changed))
	if firstHash != secondHash || firstHash == changedHash {
		t.Fatalf("fingerprint hashes = %q/%q/%q, want normalized equality and address conflict", firstHash, secondHash, changedHash)
	}
}

func TestGetOrderUsecaseEnforcesOwnershipUnlessPrivileged(t *testing.T) {
	reader := &fakeOrderReader{order: domain.Order{OrderID: "ord_1", UserID: "owner"}}
	get, _ := NewGetOrderUsecase(reader)
	if _, err := get.Execute(context.Background(), GetOrderQuery{ActorID: "other", OrderID: "ord_1", Roles: []string{"buyer"}}); !errors.Is(err, domain.ErrOrderForbidden) {
		t.Fatalf("Execute() error = %v, want ErrOrderForbidden", err)
	}
	if _, err := get.Execute(context.Background(), GetOrderQuery{ActorID: "admin_1", OrderID: "ord_1", Roles: []string{"admin"}}); err != nil {
		t.Fatalf("admin Execute() error = %v", err)
	}
}

func TestListOrdersUsecaseSignsAndValidatesCursor(t *testing.T) {
	now := paymentTestTime()
	reader := &fakeOrderReader{page: domain.OrderPage{
		Orders:     []domain.Order{{OrderID: "ord_2"}},
		NextCursor: &domain.OrderCursor{CreatedAt: now, OrderID: "ord_2"},
	}}
	list, err := NewListOrdersUsecase(reader, []byte("a pagination signing key that is secure"))
	if err != nil {
		t.Fatalf("NewListOrdersUsecase() error = %v", err)
	}
	page, err := list.Execute(context.Background(), ListOrdersQuery{UserID: "user_1", PageSize: 20})
	if err != nil || page.NextPageToken == "" {
		t.Fatalf("Execute() page/error = %+v/%v, want a continuation token", page, err)
	}
	reader.page.NextCursor = nil
	if _, err := list.Execute(context.Background(), ListOrdersQuery{
		UserID: "user_1", PageSize: 20, PageToken: page.NextPageToken,
	}); err != nil {
		t.Fatalf("Execute(next page) error = %v", err)
	}
	if reader.filter.Cursor == nil || reader.filter.Cursor.OrderID != "ord_2" || !reader.filter.Cursor.CreatedAt.Equal(now) {
		t.Fatalf("decoded cursor = %+v, want original cursor", reader.filter.Cursor)
	}
	tampered := page.NextPageToken[:len(page.NextPageToken)-1] + "x"
	if _, err := list.Execute(context.Background(), ListOrdersQuery{UserID: "user_1", PageSize: 20, PageToken: tampered}); !errors.Is(err, domain.ErrInvalidPageToken) {
		t.Fatalf("Execute(tampered token) error = %v, want ErrInvalidPageToken", err)
	}
}

func TestUpdateFulfillmentUsecaseDelegatesLegalSellerTransition(t *testing.T) {
	now := paymentTestTime()
	reader := &fakeOrderReader{order: domain.Order{OrderID: "ord_1", Status: domain.OrderStatusPaid}}
	writer := &fakeFulfillmentRepository{order: domain.Order{OrderID: "ord_1", Status: domain.OrderStatusPacked}}
	update, err := NewUpdateFulfillmentUsecase(
		writer, reader, &sequenceIDGenerator{values: []string{"osh_1", "shp_1"}},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	if err != nil {
		t.Fatalf("NewUpdateFulfillmentUsecase() error = %v", err)
	}
	update.WithClock(fixedClock{now: now})
	order, err := update.Execute(context.Background(), UpdateFulfillmentCommand{
		ActorID: "seller_1", Roles: []string{"seller"}, OrderID: "ord_1",
		TargetStatus: domain.OrderStatusPacked, OccurredAt: now,
	})
	if err != nil || order.Status != domain.OrderStatusPacked {
		t.Fatalf("Execute() order/error = %+v/%v", order, err)
	}
	if writer.transition.RequiredSellerID != "seller_1" || writer.transition.History.FromStatus == nil ||
		*writer.transition.History.FromStatus != domain.OrderStatusPaid {
		t.Fatalf("transition = %+v, want seller-scoped paid-to-packed audit", writer.transition)
	}
}

func TestUpdateFulfillmentUsecaseRejectsInvalidJumpBeforeWrite(t *testing.T) {
	reader := &fakeOrderReader{order: domain.Order{OrderID: "ord_1", Status: domain.OrderStatusPaid}}
	writer := &fakeFulfillmentRepository{}
	update, _ := NewUpdateFulfillmentUsecase(writer, reader, &sequenceIDGenerator{values: []string{"osh_1", "shp_1"}}, nil)
	_, err := update.Execute(context.Background(), UpdateFulfillmentCommand{
		ActorID: "ops_1", Roles: []string{"admin"}, OrderID: "ord_1",
		TargetStatus: domain.OrderStatusDelivered, OccurredAt: paymentTestTime(),
	})
	if !errors.Is(err, domain.ErrInvalidOrderStatusTransition) || writer.calls != 0 {
		t.Fatalf("Execute() error/calls = %v/%d, want blocked direct delivery", err, writer.calls)
	}
}

func TestUpdateFulfillmentUsecaseCreatesDeliveredOutboxEvent(t *testing.T) {
	now := paymentTestTime()
	reader := &fakeOrderReader{order: domain.Order{OrderID: "ord_1", UserID: "user_1", Status: domain.OrderStatusShipped}}
	writer := &fakeFulfillmentRepository{order: domain.Order{OrderID: "ord_1", UserID: "user_1", Status: domain.OrderStatusDelivered}}
	update, err := NewUpdateFulfillmentUsecase(
		writer, reader, &sequenceIDGenerator{values: []string{"osh_delivered", "shp_1", "evt_delivered"}},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	if err != nil {
		t.Fatalf("NewUpdateFulfillmentUsecase() error = %v", err)
	}
	update.WithClock(fixedClock{now: now})
	_, err = update.Execute(context.Background(), UpdateFulfillmentCommand{
		ActorID: "ops_1", Roles: []string{"logistics"}, OrderID: "ord_1", TraceID: "trace_1",
		TargetStatus: domain.OrderStatusDelivered, OccurredAt: now,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if writer.event == nil || writer.event.EventType != "OrderDelivered" ||
		writer.event.AggregateID != "ord_1" || writer.event.TraceID != "trace_1" {
		t.Fatalf("event = %+v, want OrderDelivered outbox event", writer.event)
	}
}

type fakeCreateFromCart struct {
	command CreateOrderFromCartCommand
	result  *CreateOrderFromCartResult
	err     error
	calls   int
}

func (f *fakeCreateFromCart) Execute(_ context.Context, command CreateOrderFromCartCommand) (*CreateOrderFromCartResult, error) {
	f.calls++
	f.command = command
	return f.result, f.err
}

type fakePaymentInitiation struct {
	command InitiateOrderPaymentCommand
	result  *InitiateOrderPaymentResult
	err     error
	calls   int
}

func (f *fakePaymentInitiation) Execute(_ context.Context, command InitiateOrderPaymentCommand) (*InitiateOrderPaymentResult, error) {
	f.calls++
	f.command = command
	return f.result, f.err
}

type fakeOrderReader struct {
	order  domain.Order
	page   domain.OrderPage
	filter domain.ListOrdersFilter
	calls  int
}

func (f *fakeOrderReader) GetOrder(context.Context, string) (domain.Order, error) {
	f.calls++
	return f.order, nil
}

func (f *fakeOrderReader) ListOrders(_ context.Context, filter domain.ListOrdersFilter) (domain.OrderPage, error) {
	f.filter = filter
	return f.page, nil
}

type fakeFulfillmentRepository struct {
	order      domain.Order
	transition domain.FulfillmentTransition
	event      *domain.OutboxEvent
	calls      int
}

func (f *fakeFulfillmentRepository) UpdateFulfillment(_ context.Context, transition domain.FulfillmentTransition, event *domain.OutboxEvent) (domain.Order, error) {
	f.calls++
	f.transition = transition
	f.event = event
	return f.order, nil
}

type fakeIdempotencyRepository struct {
	decision      domain.ClaimDecision
	orderID       string
	claimErr      error
	completeCalls int
	failedCalls   int
}

func (f *fakeIdempotencyRepository) Claim(_ context.Context, userID string, key string, hash string, expiresAt time.Time) (domain.IdempotencyClaim, error) {
	if f.claimErr != nil {
		return domain.IdempotencyClaim{}, f.claimErr
	}
	return domain.IdempotencyClaim{
		UserID: userID, Key: key, RequestHash: hash, OrderID: f.orderID,
		Status: domain.IdempotencyStatusProcessing, Decision: f.decision, ExpiresAt: expiresAt,
	}, nil
}

func (f *fakeIdempotencyRepository) Complete(context.Context, string, string, string) error {
	f.completeCalls++
	return nil
}

func (f *fakeIdempotencyRepository) MarkFailed(context.Context, string, string) error {
	f.failedCalls++
	return nil
}

func validCreateOrderCommand() CreateOrderCommand {
	return CreateOrderCommand{
		UserID: "user_1", CartID: "cart_1", IdempotencyKey: "request_1",
		ShippingAddress: validCommand().ShippingAddress,
	}
}

func newCreateWorkflow(
	t *testing.T,
	checkout CreateFromCartExecutor,
	payment PaymentInitiationExecutor,
	reader OrderReadRepository,
	idempotency IdempotencyRepository,
) *CreateOrderUsecase {
	t.Helper()
	workflow, err := NewCreateOrderUsecase(
		checkout,
		payment,
		reader,
		idempotency,
		CreateOrderConfig{IdempotencyTTL: 24 * time.Hour},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	if err != nil {
		t.Fatalf("NewCreateOrderUsecase() error = %v", err)
	}
	workflow.WithClock(fixedClock{now: paymentTestTime()})
	return workflow
}
