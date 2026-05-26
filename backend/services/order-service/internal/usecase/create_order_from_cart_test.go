package usecase

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/order-service/internal/domain"
)

func TestCreateOrderFromCartCreatesSnapshotOrderAfterReservation(t *testing.T) {
	carts, products, orders, ids, now := validDependencies()
	var logs bytes.Buffer
	usecase := newTestUsecase(t, carts, products, orders, ids, now, slog.New(slog.NewTextHandler(&logs, nil)))

	result, err := usecase.Execute(context.Background(), validCommand())
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.Status != domain.OrderStatusCreated || result.TotalAmount != 299900 {
		t.Fatalf("result = %+v, want created order for total 299900", result)
	}
	if orders.created == nil || orders.created.Items[0].TitleSnapshot != "Fresh Product Title" ||
		orders.created.Items[0].UnitAmount != 299900 ||
		orders.created.InventoryReservationID != "res_1" ||
		orders.created.InventoryReservedUntil != now.Add(10*time.Minute) {
		t.Fatalf("persisted order did not contain fresh product snapshot: %+v", orders.created)
	}
	if products.reserveRequest.TTL != 10*time.Minute || products.reserveCalls != 1 {
		t.Fatalf("reservation request = %+v, want configured TTL and one call", products.reserveRequest)
	}
	if products.reserveRequest.IdempotencyKey != childOperationKey("order-reservation", "user_1", "attempt_1") ||
		orders.claim.Key != "attempt_1" || orders.event.EventType != "OrderCreated" ||
		orders.event.AggregateID != "ord_1" {
		t.Fatalf("reservation key/claim/event = %q/%+v/%+v, want stable child key, claim, and OrderCreated event", products.reserveRequest.IdempotencyKey, orders.claim, orders.event)
	}
	if products.releaseCalls != 0 || !bytes.Contains(logs.Bytes(), []byte("order.create.from_cart_succeeded")) {
		t.Fatalf("unexpected compensation or missing success log: releases=%d logs=%s", products.releaseCalls, logs.String())
	}
}

func TestCreateOrderFromCartRejectsEmptyCartBeforeProductCalls(t *testing.T) {
	carts, products, orders, ids, now := validDependencies()
	carts.cart.Items = nil
	usecase := newTestUsecase(t, carts, products, orders, ids, now, nil)

	_, err := usecase.Execute(context.Background(), validCommand())
	if !errors.Is(err, domain.ErrCartEmpty) {
		t.Fatalf("Execute() error = %v, want ErrCartEmpty", err)
	}
	if products.batchCalls != 0 || products.reserveCalls != 0 || orders.created != nil {
		t.Fatal("invalid cart called downstream product or repository operations")
	}
}

func TestCreateOrderFromCartRejectsOversizedIdentityBeforeCartCall(t *testing.T) {
	carts, products, orders, ids, now := validDependencies()
	usecase := newTestUsecase(t, carts, products, orders, ids, now, nil)
	command := validCommand()
	command.UserID = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"

	_, err := usecase.Execute(context.Background(), command)
	if !errors.Is(err, domain.ErrInvalidCheckoutCommand) {
		t.Fatalf("Execute() error = %v, want ErrInvalidCheckoutCommand", err)
	}
	if carts.calls != 0 || products.batchCalls != 0 || orders.created != nil {
		t.Fatal("oversized request identity reached downstream operations")
	}
}

func TestCreateOrderFromCartRejectsMismatchedCartResponse(t *testing.T) {
	carts, products, orders, ids, now := validDependencies()
	carts.cart.CartID = "different_cart"
	usecase := newTestUsecase(t, carts, products, orders, ids, now, nil)

	_, err := usecase.Execute(context.Background(), validCommand())
	if !errors.Is(err, domain.ErrCartNotFound) {
		t.Fatalf("Execute() error = %v, want ErrCartNotFound", err)
	}
	if products.batchCalls != 0 || products.reserveCalls != 0 || orders.created != nil {
		t.Fatal("mismatched cart response called product or persistence operations")
	}
}

func TestCreateOrderFromCartDoesNotReserveUnavailableProduct(t *testing.T) {
	carts, products, orders, ids, now := validDependencies()
	products.snapshots.Items[0].Published = false
	usecase := newTestUsecase(t, carts, products, orders, ids, now, nil)

	_, err := usecase.Execute(context.Background(), validCommand())
	if !errors.Is(err, domain.ErrProductUnavailable) {
		t.Fatalf("Execute() error = %v, want ErrProductUnavailable", err)
	}
	if products.reserveCalls != 0 || orders.created != nil {
		t.Fatal("unavailable product reserved inventory or persisted an order")
	}
}

func TestCreateOrderFromCartReleasesReservationWhenPersistenceFailsEvenIfRequestCancelled(t *testing.T) {
	carts, products, orders, ids, now := validDependencies()
	ctx, cancel := context.WithCancel(context.Background())
	orders.createFunc = func(context.Context, domain.Order) error {
		cancel()
		return errors.New("mysql unavailable")
	}
	usecase := newTestUsecase(t, carts, products, orders, ids, now, nil)

	_, err := usecase.Execute(ctx, validCommand())
	if !errors.Is(err, domain.ErrOrderCreateFailed) {
		t.Fatalf("Execute() error = %v, want ErrOrderCreateFailed", err)
	}
	if products.releaseCalls != 1 || products.releaseRequest.ReservationID != "res_1" ||
		products.releaseRequest.Reason != inventoryReleaseReasonOrderPersistenceFailed {
		t.Fatalf("release request = %+v calls=%d, want order persistence compensation", products.releaseRequest, products.releaseCalls)
	}
	if products.releaseContextErr != nil {
		t.Fatalf("release context inherited cancelled request: %v", products.releaseContextErr)
	}
}

func TestCreateOrderFromCartDoesNotReleaseReservationWhenCommitOutcomeIsUnknown(t *testing.T) {
	carts, products, orders, ids, now := validDependencies()
	orders.createFunc = func(context.Context, domain.Order) error {
		return domain.ErrOrderCreateOutcomeUnknown
	}
	usecase := newTestUsecase(t, carts, products, orders, ids, now, nil)

	_, err := usecase.Execute(context.Background(), validCommand())
	if !errors.Is(err, domain.ErrOrderCreateOutcomeUnknown) {
		t.Fatalf("Execute() error = %v, want ErrOrderCreateOutcomeUnknown", err)
	}
	if products.releaseCalls != 0 {
		t.Fatalf("release calls = %d, an uncertain commit must remain reserved for reconciliation", products.releaseCalls)
	}
}

func TestCreateOrderFromCartReturnsInventoryBusinessErrorWithoutWritingOrder(t *testing.T) {
	carts, products, orders, ids, now := validDependencies()
	products.reserveErr = domain.ErrInventoryUnavailable
	usecase := newTestUsecase(t, carts, products, orders, ids, now, nil)

	_, err := usecase.Execute(context.Background(), validCommand())
	if !errors.Is(err, domain.ErrInventoryUnavailable) {
		t.Fatalf("Execute() error = %v, want ErrInventoryUnavailable", err)
	}
	if orders.created != nil {
		t.Fatal("inventory failure persisted an order")
	}
}

type fakeCartClient struct {
	cart  *domain.CartSnapshot
	calls int
}

func (f *fakeCartClient) GetCart(context.Context, GetCartRequest) (*domain.CartSnapshot, error) {
	f.calls++
	return f.cart, nil
}

type fakeProductClient struct {
	snapshots         *BatchGetProductsResponse
	reservation       *ReserveInventoryResponse
	reserveErr        error
	batchCalls        int
	reserveCalls      int
	releaseCalls      int
	reserveRequest    ReserveInventoryRequest
	releaseRequest    ReleaseInventoryRequest
	releaseContextErr error
}

func (f *fakeProductClient) BatchGetProducts(context.Context, BatchGetProductsRequest) (*BatchGetProductsResponse, error) {
	f.batchCalls++
	return f.snapshots, nil
}

func (f *fakeProductClient) ReserveInventory(_ context.Context, request ReserveInventoryRequest) (*ReserveInventoryResponse, error) {
	f.reserveCalls++
	f.reserveRequest = request
	return f.reservation, f.reserveErr
}

func (f *fakeProductClient) ReleaseInventory(ctx context.Context, request ReleaseInventoryRequest) error {
	f.releaseCalls++
	f.releaseRequest = request
	f.releaseContextErr = ctx.Err()
	return nil
}

type fakeOrderRepository struct {
	created    *domain.Order
	claim      domain.IdempotencyClaim
	event      domain.OutboxEvent
	createFunc func(context.Context, domain.Order) error
}

func (f *fakeOrderRepository) CreateOrderAndAttachClaim(ctx context.Context, order domain.Order, claim domain.IdempotencyClaim, event domain.OutboxEvent) error {
	copy := order
	f.created = &copy
	f.claim = claim
	f.event = event
	if f.createFunc != nil {
		return f.createFunc(ctx, order)
	}
	return nil
}

type sequenceIDGenerator struct {
	values []string
}

func (g *sequenceIDGenerator) NewID(string) (string, error) {
	value := g.values[0]
	g.values = g.values[1:]
	return value, nil
}

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time {
	return c.now
}

func validDependencies() (*fakeCartClient, *fakeProductClient, *fakeOrderRepository, *sequenceIDGenerator, time.Time) {
	now := time.Date(2026, time.May, 26, 10, 0, 0, 0, time.UTC)
	carts := &fakeCartClient{cart: &domain.CartSnapshot{
		CartID: "cart_1", UserID: "user_1", Currency: "INR",
		Items: []domain.CartItemSnapshot{{ProductID: "product_1", VariantID: "variant_1", Quantity: 1}},
	}}
	products := &fakeProductClient{
		snapshots: &BatchGetProductsResponse{Items: []domain.ProductSnapshot{{
			ProductID: "product_1", VariantID: "variant_1", SellerID: "seller_1", SKU: "SKU-1",
			Title: "Fresh Product Title", Currency: "INR", UnitAmount: 299900,
			Published: true, VariantActive: true, InStock: true,
		}}},
		reservation: &ReserveInventoryResponse{ReservationID: "res_1", ExpiresAt: now.Add(10 * time.Minute)},
	}
	return carts, products, &fakeOrderRepository{}, &sequenceIDGenerator{values: []string{"ord_1", "osh_1", "oi_1", "evt_created"}}, now
}

func validCommand() CreateOrderFromCartCommand {
	return CreateOrderFromCartCommand{
		UserID: "user_1", CartID: "cart_1", SessionID: "session_1", IdempotencyKey: "attempt_1",
		ShippingAddress: domain.AddressSnapshot{
			RecipientName: "Buyer", Line1: "101 Main Street", City: "Delhi", PostalCode: "110001", Country: "IN",
		},
		Claim: domain.IdempotencyClaim{
			UserID: "user_1", Key: "attempt_1",
			RequestHash: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			Status:      domain.IdempotencyStatusProcessing, Decision: domain.ClaimAcquired,
			ExpiresAt: time.Date(2026, time.May, 27, 10, 0, 0, 0, time.UTC),
		},
	}
}

func newTestUsecase(
	t *testing.T,
	carts CartClient,
	products ProductClient,
	orders OrderRepository,
	ids IDGenerator,
	now time.Time,
	logger *slog.Logger,
) *CreateOrderFromCartUsecase {
	t.Helper()
	usecase, err := NewCreateOrderFromCartUsecase(carts, products, orders, ids, CreateOrderFromCartConfig{
		ReservationTTL: 10 * time.Minute,
		ReleaseTimeout: time.Second,
	}, logger)
	if err != nil {
		t.Fatalf("NewCreateOrderFromCartUsecase() error = %v", err)
	}
	usecase.WithClock(fixedClock{now: now})
	return usecase
}
