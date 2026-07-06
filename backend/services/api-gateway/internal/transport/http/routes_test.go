package httptransport

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	gatewayauth "ecommerce/api-gateway/internal/auth"
	"ecommerce/api-gateway/internal/clients"
	"ecommerce/api-gateway/internal/config"
	"ecommerce/api-gateway/internal/repository"
	"ecommerce/api-gateway/internal/usecase"
	"github.com/golang-jwt/jwt/v5"
	notificationv1 "github.com/parag/ecommerce/backend/shared/gen/go/ecommerce/notification/v1"
	orderv1 "github.com/parag/ecommerce/backend/shared/gen/go/ecommerce/order/v1"
	userv1 "github.com/parag/ecommerce/backend/shared/gen/go/ecommerce/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestNewRouterRegistersDefinedRoute(t *testing.T) {
	router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/products/p_123", nil)
	req.Header.Set("X-Request-Id", "req_test")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501 for defined route bridge, got %d", rec.Code)
	}
	var envelope ResponseEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if envelope.RequestID != "req_test" {
		t.Fatalf("expected request id propagation, got %q", envelope.RequestID)
	}
	if envelope.Error == nil || envelope.Error.Code != "ROUTE_BRIDGE_NOT_CONFIGURED" {
		t.Fatalf("expected bridge error envelope, got %+v", envelope.Error)
	}
}

func TestNewRouterHealthReady(t *testing.T) {
	router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected ready status 200, got %d", rec.Code)
	}
	var envelope ResponseEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	data, ok := envelope.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected health data map, got %T", envelope.Data)
	}
	if data["status"] != "ready" {
		t.Fatalf("expected ready status data, got %+v", data)
	}
}

func TestNewRouterHealthReadyChecksDownstreams(t *testing.T) {
	router := newTestRouterWithHealth(t, stubDownstreamHealth{
		report: clients.HealthReport{
			clients.DownstreamAuth: {
				Service: "auth",
				Status:  clients.HealthStatusUnavailable,
				Error:   "connection refused",
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected readiness failure 503, got %d", rec.Code)
	}
	var envelope ResponseEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if envelope.Error == nil || envelope.Error.Code != "DOWNSTREAM_SERVICES_UNAVAILABLE" {
		t.Fatalf("expected downstream readiness error, got %+v", envelope.Error)
	}
	details, ok := envelope.Error.Details.([]any)
	if !ok || len(details) != 1 {
		t.Fatalf("expected readiness error details, got %#v", envelope.Error.Details)
	}
	detail, ok := details[0].(map[string]any)
	if !ok || detail["reason"] != "auth: connection refused" {
		t.Fatalf("expected downstream error reason, got %#v", details[0])
	}
}

func TestProtectedRouteRequiresBearerToken(t *testing.T) {
	router := newAuthTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing token, got %d", rec.Code)
	}
	assertErrorCode(t, rec, "UNAUTHORIZED")
}

func TestProtectedRouteAllowsMatchingRole(t *testing.T) {
	router := newAuthTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer buyer-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("expected route bridge response after auth, got %d", rec.Code)
	}
	assertErrorCode(t, rec, "ROUTE_BRIDGE_NOT_CONFIGURED")
}

func TestProtectedUserRouteBridgesAuthenticatedIdentityToGRPC(t *testing.T) {
	path := writeAuthTestContract(t)
	cfg := authTestConfig(path)
	repo := repository.NewJSONRouteRepository(path)
	catalog := usecase.NewRouteCatalogService(repo, cfg.APIBasePath)
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	userClient := &routerUserClient{
		getUser: func(ctx context.Context, request *userv1.GetUserRequest) (*userv1.UserProfile, error) {
			if request.GetUserId() != "user_buyer" {
				t.Fatalf("user id = %q, want user_buyer", request.GetUserId())
			}
			outgoing, ok := metadata.FromOutgoingContext(ctx)
			if !ok || len(outgoing.Get("x-user-id")) != 1 || outgoing.Get("x-user-id")[0] != "user_buyer" {
				t.Fatalf("outgoing x-user-id = %#v", outgoing.Get("x-user-id"))
			}
			return &userv1.UserProfile{UserId: request.GetUserId(), Email: "buyer@example.com"}, nil
		},
	}
	router, err := NewRouterWithOptions(context.Background(), cfg, catalog, logger, nil, RouterOptions{
		TokenVerifier: stubTokenVerifier{},
		UserClient:    userClient,
	})
	if err != nil {
		t.Fatalf("new router: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer buyer-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected bridged profile response 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var envelope ResponseEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	data, ok := envelope.Data.(map[string]any)
	if !ok || data["user_id"] != "user_buyer" {
		t.Fatalf("unexpected profile response: %#v", envelope.Data)
	}
}

func TestProtectedWishlistRouteProxiesAuthenticatedIdentityToHTTP(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-User-ID"); got != "user_buyer" {
			t.Fatalf("upstream X-User-ID = %q, want user_buyer", got)
		}
		if got := r.Header.Get("X-User-Roles"); got != "buyer" {
			t.Fatalf("upstream X-User-Roles = %q, want buyer", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"items":[]},"request_id":"req_upstream","error":null}`))
	}))
	defer upstream.Close()

	path := filepath.Join(t.TempDir(), "master-api.json")
	contract := `{"project":"test","version":"1.0.0","rest_endpoints":[{"id":"wishlist.get","method":"GET","path":"/api/v1/wishlist","service":"wishlist-service","grpc":"WishlistService.GetWishlist","auth":"buyer","request_schema":"Empty","response_schema":"Wishlist"}]}`
	if err := os.WriteFile(path, []byte(contract), 0o600); err != nil {
		t.Fatalf("write contract: %v", err)
	}
	cfg := authTestConfig(path)
	cfg.WishlistHTTPURL = upstream.URL
	cfg.WishlistHTTPTimeout = time.Second
	repo := repository.NewJSONRouteRepository(path)
	catalog := usecase.NewRouteCatalogService(repo, cfg.APIBasePath)
	router, err := NewRouterWithOptions(context.Background(), cfg, catalog, slog.New(slog.NewJSONHandler(io.Discard, nil)), nil, RouterOptions{
		TokenVerifier:      stubTokenVerifier{},
		WishlistHTTPClient: upstream.Client(),
	})
	if err != nil {
		t.Fatalf("new router: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/wishlist", nil)
	request.Header.Set("Authorization", "Bearer buyer-token")
	request.Header.Set("X-User-ID", "spoofed")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 body=%s", response.Code, response.Body.String())
	}
}

func TestProtectedCartRouteProxiesAuthenticatedIdentityToHTTP(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-User-ID"); got != "user_buyer" {
			t.Fatalf("upstream X-User-ID = %q, want user_buyer", got)
		}
		if got := r.Header.Get("X-User-Roles"); got != "buyer" {
			t.Fatalf("upstream X-User-Roles = %q, want buyer", got)
		}
		if got := r.Header.Get("X-Session-ID"); got != "sess_buyer" {
			t.Fatalf("upstream X-Session-ID = %q, want sess_buyer", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"cart_id":"cart_1","items":[]}`))
	}))
	defer upstream.Close()

	path := filepath.Join(t.TempDir(), "master-api.json")
	contract := `{"project":"test","version":"1.0.0","rest_endpoints":[{"id":"cart.get","method":"GET","path":"/api/v1/cart","service":"cart-service","grpc":"CartService.GetCart","auth":"buyer","request_schema":"Empty","response_schema":"Cart"}]}`
	if err := os.WriteFile(path, []byte(contract), 0o600); err != nil {
		t.Fatalf("write contract: %v", err)
	}
	cfg := authTestConfig(path)
	cfg.CartHTTPURL = upstream.URL
	cfg.CartHTTPTimeout = time.Second
	repo := repository.NewJSONRouteRepository(path)
	catalog := usecase.NewRouteCatalogService(repo, cfg.APIBasePath)
	router, err := NewRouterWithOptions(context.Background(), cfg, catalog, slog.New(slog.NewJSONHandler(io.Discard, nil)), nil, RouterOptions{
		TokenVerifier:  stubTokenVerifier{},
		CartHTTPClient: upstream.Client(),
	})
	if err != nil {
		t.Fatalf("new router: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/cart", nil)
	request.Header.Set("Authorization", "Bearer buyer-token")
	request.Header.Set("X-User-ID", "spoofed")
	request.Header.Set("X-Session-ID", "spoofed_session")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 body=%s", response.Code, response.Body.String())
	}
}

func TestAnalyticsRouteProxiesWithInternalSessionAuth(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/analytics/live" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer session-service-admin-token-at-least-32-chars" {
			t.Fatalf("authorization=%q", r.Header.Get("Authorization"))
		}
		if r.Header.Get("X-User-ID") != "admin_1" || r.Header.Get("X-User-Roles") != "operations_admin" {
			t.Fatalf("trusted identity headers=%v", r.Header)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"active_users":1,"active_sessions":1,"events_per_minute":0}`)
	}))
	defer upstream.Close()

	path := filepath.Join(t.TempDir(), "master-api.json")
	contract := `{"project":"test","version":"1.0.0","rest_endpoints":[{"id":"analytics.live","method":"GET","path":"/api/v1/analytics/live","service":"session-service","grpc":"SessionService.GetLiveMetrics","auth":"admin","request_schema":"LiveMetricsRequest","response_schema":"LiveMetricsResponse"}]}`
	if err := os.WriteFile(path, []byte(contract), 0o600); err != nil {
		t.Fatalf("write contract: %v", err)
	}
	cfg := authTestConfig(path)
	cfg.SessionHTTPURL = upstream.URL
	cfg.SessionHTTPTimeout = time.Second
	cfg.SessionServiceAdminToken = "session-service-admin-token-at-least-32-chars"
	repo := repository.NewJSONRouteRepository(path)
	catalog := usecase.NewRouteCatalogService(repo, cfg.APIBasePath)
	router, err := NewRouterWithOptions(context.Background(), cfg, catalog, slog.New(slog.NewJSONHandler(io.Discard, nil)), nil, RouterOptions{
		TokenVerifier:     stubTokenVerifier{},
		SessionHTTPClient: upstream.Client(),
	})
	if err != nil {
		t.Fatalf("new router: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/live", nil)
	request.Header.Set("Authorization", "Bearer admin-token")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 body=%s", response.Code, response.Body.String())
	}
}

func TestBuyerCheckoutRouteBridgesToOrderService(t *testing.T) {
	path := filepath.Join(t.TempDir(), "master-api.json")
	contract := `{"project":"test","version":"1.0.0","rest_endpoints":[{"id":"order.checkout","method":"POST","path":"/api/v1/orders/checkout","service":"order-service","grpc":"OrderService.CreateOrder","auth":"buyer","request_schema":"CheckoutRequest","response_schema":"CheckoutResponse"}],"schemas":{"CheckoutRequest":{"type":"object","required":["address_id","cart_id","payment_provider","idempotency_key"],"properties":{"address_id":{"type":"string"},"cart_id":{"type":"string"},"payment_provider":{"type":"string"},"idempotency_key":{"type":"string"}}}}}`
	if err := os.WriteFile(path, []byte(contract), 0o600); err != nil {
		t.Fatalf("write contract: %v", err)
	}
	cfg := authTestConfig(path)
	repo := repository.NewJSONRouteRepository(path)
	catalog := usecase.NewRouteCatalogService(repo, cfg.APIBasePath)
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	now := time.Date(2026, 7, 2, 12, 0, 0, 0, time.UTC)
	userClient := &routerUserClient{
		listAddresses: func(ctx context.Context, request *userv1.ListUserAddressesRequest) (*userv1.AddressListResponse, error) {
			if request.GetUserId() != "user_buyer" {
				t.Fatalf("address user_id = %q, want user_buyer", request.GetUserId())
			}
			if request.GetPageSize() != 100 {
				t.Fatalf("address page_size = %d, want 100", request.GetPageSize())
			}
			outgoing, ok := metadata.FromOutgoingContext(ctx)
			if !ok || len(outgoing.Get("x-user-id")) != 1 || outgoing.Get("x-user-id")[0] != "user_buyer" {
				t.Fatalf("address outgoing x-user-id = %#v", outgoing.Get("x-user-id"))
			}
			return &userv1.AddressListResponse{
				Addresses: []*userv1.Address{{
					AddressId:  "addr_1",
					Name:       "Parag",
					Phone:      "+919990494800",
					Line1:      "1E/37",
					City:       "Faridabad",
					State:      "Haryana",
					PostalCode: "121001",
					Country:    "India",
				}},
				Page:     1,
				PageSize: 100,
			}, nil
		},
	}
	orderClient := &routerOrderClient{
		createOrder: func(ctx context.Context, request *orderv1.CreateOrderRequest, _ ...grpc.CallOption) (*orderv1.CreateOrderResponse, error) {
			if request.GetCartId() != "cart_1" {
				t.Fatalf("cart_id = %q, want cart_1", request.GetCartId())
			}
			if request.GetIdempotencyKey() != "idem_123456789" {
				t.Fatalf("idempotency_key = %q, want idem_123456789", request.GetIdempotencyKey())
			}
			if request.GetShippingAddress().GetCountryCode() != "IN" {
				t.Fatalf("country_code = %q, want IN", request.GetShippingAddress().GetCountryCode())
			}
			outgoing, ok := metadata.FromOutgoingContext(ctx)
			if !ok || len(outgoing.Get("x-user-id")) != 1 || outgoing.Get("x-user-id")[0] != "user_buyer" {
				t.Fatalf("order outgoing x-user-id = %#v", outgoing.Get("x-user-id"))
			}
			return &orderv1.CreateOrderResponse{
				Order: &orderv1.Order{
					OrderId:         "ord_1",
					UserId:          "user_buyer",
					Status:          orderv1.OrderStatus_ORDER_STATUS_PENDING_PAYMENT,
					Total:           &orderv1.Money{MinorUnits: 499, Currency: "INR"},
					ShippingAddress: request.GetShippingAddress(),
					PaymentId:       "pay_1",
					CreatedAt:       timestamppb.New(now),
					UpdatedAt:       timestamppb.New(now),
					Items: []*orderv1.OrderItem{{
						OrderItemId: "item_1",
						ProductId:   "prod_1",
						VariantId:   "var_1",
						ProductName: "Cotton T-shirt",
						Quantity:    1,
						UnitPrice:   &orderv1.Money{MinorUnits: 499, Currency: "INR"},
						LineTotal:   &orderv1.Money{MinorUnits: 499, Currency: "INR"},
					}},
				},
				PaymentAction: &orderv1.PaymentAction{
					PaymentId:         "pay_1",
					ClientActionToken: "client_secret_1",
					ExpiresAt:         timestamppb.New(now.Add(15 * time.Minute)),
				},
			}, nil
		},
	}
	router, err := NewRouterWithOptions(context.Background(), cfg, catalog, logger, nil, RouterOptions{
		TokenVerifier: stubTokenVerifier{},
		UserClient:    userClient,
		OrderClient:   orderClient,
	})
	if err != nil {
		t.Fatalf("new router: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/v1/orders/checkout", strings.NewReader(`{"address_id":"addr_1","cart_id":"cart_1","payment_provider":"razorpay","idempotency_key":"idem_123456789"}`))
	request.Header.Set("Authorization", "Bearer buyer-token")
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 body=%s", response.Code, response.Body.String())
	}
	var envelope ResponseEnvelope
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	data, ok := envelope.Data.(map[string]any)
	if !ok {
		t.Fatalf("unexpected checkout response: %#v", envelope.Data)
	}
	order, ok := data["order"].(map[string]any)
	if !ok || order["order_id"] != "ord_1" {
		t.Fatalf("unexpected order payload: %#v", data["order"])
	}
	paymentIntent, ok := data["payment_intent"].(map[string]any)
	if !ok || paymentIntent["payment_id"] != "pay_1" || paymentIntent["status"] != "requires_action" {
		t.Fatalf("unexpected payment intent: %#v", data["payment_intent"])
	}
}

func TestSellerSessionRouteBuildsFromAuthenticatedClaims(t *testing.T) {
	path := filepath.Join(t.TempDir(), "master-api.json")
	contract := `{"project":"test","version":"1.0.0","rest_endpoints":[{"id":"seller.session_get","method":"GET","path":"/api/v1/seller/session","service":"api-gateway-service","grpc":"GatewayService.GetSellerSession","auth":"seller","request_schema":"Empty","response_schema":"SellerSessionResponse"}]}`
	if err := os.WriteFile(path, []byte(contract), 0o600); err != nil {
		t.Fatalf("write contract: %v", err)
	}
	cfg := authTestConfig(path)
	repo := repository.NewJSONRouteRepository(path)
	catalog := usecase.NewRouteCatalogService(repo, cfg.APIBasePath)
	router, err := NewRouterWithOptions(context.Background(), cfg, catalog, slog.New(slog.NewJSONHandler(io.Discard, nil)), nil, RouterOptions{
		TokenVerifier: stubTokenVerifier{},
	})
	if err != nil {
		t.Fatalf("new router: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/seller/session", nil)
	request.Header.Set("Authorization", "Bearer seller-token")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 body=%s", response.Code, response.Body.String())
	}
	var envelope ResponseEnvelope
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	data, ok := envelope.Data.(map[string]any)
	if !ok || data["authenticated"] != true {
		t.Fatalf("unexpected session response: %#v", envelope.Data)
	}
	activeSeller, ok := data["active_seller"].(map[string]any)
	if !ok || activeSeller["seller_id"] != "seller_123" {
		t.Fatalf("unexpected active seller: %#v", data["active_seller"])
	}
}

func TestProductRouteProxiesSellerIdentityToHTTP(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-User-ID"); got != "user_seller" {
			t.Fatalf("upstream X-User-ID = %q, want user_seller", got)
		}
		if got := r.Header.Get("X-Seller-ID"); got != "seller_123" {
			t.Fatalf("upstream X-Seller-ID = %q, want seller_123", got)
		}
		if got := r.Header.Get("X-Roles"); got != "seller" {
			t.Fatalf("upstream X-Roles = %q, want seller", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"product_id":"prod_1"}`))
	}))
	defer upstream.Close()

	path := filepath.Join(t.TempDir(), "master-api.json")
	contract := `{"project":"test","version":"1.0.0","rest_endpoints":[{"id":"seller.product_create","method":"POST","path":"/api/v1/seller/products","service":"product-service","grpc":"ProductService.CreateProduct","auth":"seller","request_schema":"ProductInput","response_schema":"Product"}]}`
	if err := os.WriteFile(path, []byte(contract), 0o600); err != nil {
		t.Fatalf("write contract: %v", err)
	}
	cfg := authTestConfig(path)
	cfg.ProductHTTPURL = upstream.URL
	cfg.ProductHTTPTimeout = time.Second
	repo := repository.NewJSONRouteRepository(path)
	catalog := usecase.NewRouteCatalogService(repo, cfg.APIBasePath)
	router, err := NewRouterWithOptions(context.Background(), cfg, catalog, slog.New(slog.NewJSONHandler(io.Discard, nil)), nil, RouterOptions{
		TokenVerifier:     stubTokenVerifier{},
		ProductHTTPClient: upstream.Client(),
	})
	if err != nil {
		t.Fatalf("new router: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/v1/seller/products", strings.NewReader(`{}`))
	request.Header.Set("Authorization", "Bearer seller-token")
	request.Header.Set("X-Seller-ID", "spoofed")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201 body=%s", response.Code, response.Body.String())
	}
}

func TestCMSRouteProxiesSellerIdentityAndInternalAuthToHTTP(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Internal-Token"); got != "cms-secret" {
			t.Fatalf("upstream X-Internal-Token = %q, want cms-secret", got)
		}
		if got := r.Header.Get("X-Seller-ID"); got != "seller_123" {
			t.Fatalf("upstream X-Seller-ID = %q, want seller_123", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"coupons":[]}`))
	}))
	defer upstream.Close()

	path := filepath.Join(t.TempDir(), "master-api.json")
	contract := `{"project":"test","version":"1.0.0","rest_endpoints":[{"id":"cms.coupon_list","method":"GET","path":"/api/v1/seller/coupons","service":"cms-service","grpc":"CMSService.ListCoupons","auth":"seller","request_schema":"PaginationRequest","response_schema":"CouponListResponse"}]}`
	if err := os.WriteFile(path, []byte(contract), 0o600); err != nil {
		t.Fatalf("write contract: %v", err)
	}
	cfg := authTestConfig(path)
	cfg.CMSHTTPURL = upstream.URL
	cfg.CMSHTTPTimeout = time.Second
	cfg.CMSInternalAuthHeader = "X-Internal-Token"
	cfg.CMSInternalAuthToken = "cms-secret"
	repo := repository.NewJSONRouteRepository(path)
	catalog := usecase.NewRouteCatalogService(repo, cfg.APIBasePath)
	router, err := NewRouterWithOptions(context.Background(), cfg, catalog, slog.New(slog.NewJSONHandler(io.Discard, nil)), nil, RouterOptions{
		TokenVerifier: stubTokenVerifier{},
		CMSHTTPClient: upstream.Client(),
	})
	if err != nil {
		t.Fatalf("new router: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/seller/coupons", nil)
	request.Header.Set("Authorization", "Bearer seller-token")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 body=%s", response.Code, response.Body.String())
	}
}

func TestSellerOrderRouteBridgesAuthenticatedIdentityToGRPC(t *testing.T) {
	path := filepath.Join(t.TempDir(), "master-api.json")
	contract := `{"project":"test","version":"1.0.0","rest_endpoints":[{"id":"seller.order.list","method":"GET","path":"/api/v1/seller/orders","service":"order-service","grpc":"OrderService.ListSellerOrders","auth":"seller","request_schema":"SellerOrderListRequest","response_schema":"OrderListResponse"}],"schemas":{"PaginationRequest":{"type":"object","properties":{"page":{"type":"integer","minimum":1},"page_size":{"type":"integer","minimum":1,"maximum":100},"cursor":{"type":"string"}}},"SellerOrderListRequest":{"allOf":[{"$ref":"#/schemas/PaginationRequest"},{"type":"object","properties":{"status":{"type":"string"},"q":{"type":"string"},"date_from":{"type":"string"},"date_to":{"type":"string"},"page_token":{"type":"string"}}}]}}}`
	if err := os.WriteFile(path, []byte(contract), 0o600); err != nil {
		t.Fatalf("write contract: %v", err)
	}
	cfg := authTestConfig(path)
	repo := repository.NewJSONRouteRepository(path)
	catalog := usecase.NewRouteCatalogService(repo, cfg.APIBasePath)
	orderClient := &routerOrderClient{
		listSellerOrders: func(ctx context.Context, request *orderv1.ListSellerOrdersRequest, _ ...grpc.CallOption) (*orderv1.ListSellerOrdersResponse, error) {
			if request.GetPageSize() != 25 {
				t.Fatalf("page_size = %d, want 25", request.GetPageSize())
			}
			if request.GetPageToken() != "cursor_1" {
				t.Fatalf("page_token = %q, want cursor_1", request.GetPageToken())
			}
			if request.GetFulfillmentFilter() != orderv1.SellerFulfillmentStatus_SELLER_FULFILLMENT_STATUS_SHIPPED {
				t.Fatalf("fulfillment_filter = %v, want shipped", request.GetFulfillmentFilter())
			}
			outgoing, ok := metadata.FromOutgoingContext(ctx)
			if !ok || len(outgoing.Get("x-seller-id")) != 1 || outgoing.Get("x-seller-id")[0] != "seller_123" {
				t.Fatalf("outgoing x-seller-id = %#v", outgoing.Get("x-seller-id"))
			}
			if got := outgoing.Get("x-user-id"); len(got) != 1 || got[0] != "user_seller" {
				t.Fatalf("outgoing x-user-id = %#v", got)
			}
			return &orderv1.ListSellerOrdersResponse{
				Orders: []*orderv1.SellerOrderView{{
					OrderId:                 "ord_1",
					ParentOrderStatus:       orderv1.OrderStatus_ORDER_STATUS_PAID,
					SellerFulfillmentStatus: orderv1.SellerFulfillmentStatus_SELLER_FULFILLMENT_STATUS_SHIPPED,
					SellerItemsTotal:        &orderv1.Money{MinorUnits: 129900, Currency: "INR"},
					CreatedAt:               timestamppb.New(time.Date(2026, 6, 25, 10, 30, 0, 0, time.UTC)),
					Items: []*orderv1.SellerOrderItem{{
						OrderItemId:       "oi_1",
						ProductId:         "prod_1",
						Sku:               "SKU-1",
						TitleSnapshot:     "Running Shoe",
						Quantity:          1,
						UnitPrice:         &orderv1.Money{MinorUnits: 129900, Currency: "INR"},
						LineTotal:         &orderv1.Money{MinorUnits: 129900, Currency: "INR"},
						FulfillmentStatus: orderv1.ItemFulfillmentStatus_ITEM_FULFILLMENT_STATUS_SHIPPED,
					}},
				}},
				NextPageToken: "cursor_2",
			}, nil
		},
	}
	router, err := NewRouterWithOptions(context.Background(), cfg, catalog, slog.New(slog.NewJSONHandler(io.Discard, nil)), nil, RouterOptions{
		TokenVerifier: stubTokenVerifier{},
		OrderClient:   orderClient,
	})
	if err != nil {
		t.Fatalf("new router: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/seller/orders?status=shipped&page_size=25&cursor=cursor_1", nil)
	request.Header.Set("Authorization", "Bearer seller-token")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 body=%s", response.Code, response.Body.String())
	}
	var envelope ResponseEnvelope
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	data, ok := envelope.Data.(map[string]any)
	if !ok || data["total"].(float64) != 1 || data["next_page_token"] != "cursor_2" {
		t.Fatalf("unexpected list data: %#v", envelope.Data)
	}
	orders, ok := data["orders"].([]any)
	if !ok || len(orders) != 1 {
		t.Fatalf("unexpected orders: %#v", data["orders"])
	}
	order := orders[0].(map[string]any)
	if order["order_id"] != "ord_1" || order["status"] != "shipped" {
		t.Fatalf("unexpected order response: %#v", order)
	}
}

func TestSellerOrderFulfillmentRouteBridgesUpdateToGRPC(t *testing.T) {
	path := filepath.Join(t.TempDir(), "master-api.json")
	contract := `{"project":"test","version":"1.0.0","rest_endpoints":[{"id":"seller.order.fulfillment","method":"PATCH","path":"/api/v1/seller/orders/{order_id}/fulfillment","service":"order-service","grpc":"OrderService.UpdateFulfillment","auth":"seller","request_schema":"FulfillmentUpdateRequest","response_schema":"Order"}],"schemas":{"FulfillmentUpdateRequest":{"type":"object","properties":{"order_id":{"type":"string"},"status":{"type":"string"},"tracking_number":{"type":"string"},"carrier":{"type":"string"}}}}}`
	if err := os.WriteFile(path, []byte(contract), 0o600); err != nil {
		t.Fatalf("write contract: %v", err)
	}
	cfg := authTestConfig(path)
	repo := repository.NewJSONRouteRepository(path)
	catalog := usecase.NewRouteCatalogService(repo, cfg.APIBasePath)
	orderClient := &routerOrderClient{
		updateFulfillment: func(ctx context.Context, request *orderv1.UpdateFulfillmentRequest, _ ...grpc.CallOption) (*orderv1.UpdateFulfillmentResponse, error) {
			if request.GetOrderId() != "ord_1" || request.GetTargetStatus() != orderv1.OrderStatus_ORDER_STATUS_SHIPPED {
				t.Fatalf("unexpected update request: %#v", request)
			}
			if request.GetTrackingNumber() != "TRK123" || request.GetCarrier() != "Delhivery" {
				t.Fatalf("unexpected tracking request: %#v", request)
			}
			outgoing, ok := metadata.FromOutgoingContext(ctx)
			if !ok || len(outgoing.Get("x-seller-id")) != 1 || outgoing.Get("x-seller-id")[0] != "seller_123" {
				t.Fatalf("outgoing x-seller-id = %#v", outgoing.Get("x-seller-id"))
			}
			return &orderv1.UpdateFulfillmentResponse{
				SellerOrder: &orderv1.SellerOrderView{
					OrderId:                 "ord_1",
					ParentOrderStatus:       orderv1.OrderStatus_ORDER_STATUS_PAID,
					SellerFulfillmentStatus: orderv1.SellerFulfillmentStatus_SELLER_FULFILLMENT_STATUS_SHIPPED,
					SellerItemsTotal:        &orderv1.Money{MinorUnits: 50000, Currency: "INR"},
				},
			}, nil
		},
	}
	router, err := NewRouterWithOptions(context.Background(), cfg, catalog, slog.New(slog.NewJSONHandler(io.Discard, nil)), nil, RouterOptions{
		TokenVerifier: stubTokenVerifier{},
		OrderClient:   orderClient,
	})
	if err != nil {
		t.Fatalf("new router: %v", err)
	}

	request := httptest.NewRequest(http.MethodPatch, "/api/v1/seller/orders/ord_1/fulfillment", strings.NewReader(`{"order_id":"ord_1","status":"shipped","tracking_number":"TRK123","carrier":"Delhivery"}`))
	request.Header.Set("Authorization", "Bearer seller-token")
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 body=%s", response.Code, response.Body.String())
	}
	var envelope ResponseEnvelope
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	data, ok := envelope.Data.(map[string]any)
	if !ok || data["order_id"] != "ord_1" || data["status"] != "shipped" {
		t.Fatalf("unexpected update response: %#v", envelope.Data)
	}
}

func TestProtectedNotificationPreferenceRoutesBridgeAuthenticatedIdentityToGRPC(t *testing.T) {
	path := writeAuthTestContract(t)
	cfg := authTestConfig(path)
	repo := repository.NewJSONRouteRepository(path)
	catalog := usecase.NewRouteCatalogService(repo, cfg.APIBasePath)
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	client := &routerNotificationClient{
		get: func(ctx context.Context, _ *notificationv1.GetNotificationPreferenceRequest) (*notificationv1.NotificationPreference, error) {
			assertNotificationMetadata(t, ctx)
			return &notificationv1.NotificationPreference{EmailEnabled: true}, nil
		},
		update: func(ctx context.Context, request *notificationv1.UpdateNotificationPreferenceRequest) (*notificationv1.NotificationPreference, error) {
			assertNotificationMetadata(t, ctx)
			if request.MarketingEnabled == nil || !request.GetMarketingEnabled() {
				t.Fatalf("marketing_enabled = %#v, want true", request.MarketingEnabled)
			}
			return &notificationv1.NotificationPreference{EmailEnabled: true, MarketingEnabled: true}, nil
		},
	}
	router, err := NewRouterWithOptions(context.Background(), cfg, catalog, logger, nil, RouterOptions{
		TokenVerifier:      stubTokenVerifier{},
		NotificationClient: client,
	})
	if err != nil {
		t.Fatalf("new router: %v", err)
	}

	getRequest := httptest.NewRequest(http.MethodGet, "/api/v1/me/notification-preferences", nil)
	getRequest.Header.Set("Authorization", "Bearer buyer-token")
	getResponse := httptest.NewRecorder()
	router.ServeHTTP(getResponse, getRequest)
	if getResponse.Code != http.StatusOK {
		t.Fatalf("GET status = %d: %s", getResponse.Code, getResponse.Body.String())
	}

	patchRequest := httptest.NewRequest(http.MethodPatch, "/api/v1/me/notification-preferences", strings.NewReader(`{"marketing_enabled":true}`))
	patchRequest.Header.Set("Authorization", "Bearer buyer-token")
	patchRequest.Header.Set("Content-Type", "application/json")
	patchResponse := httptest.NewRecorder()
	router.ServeHTTP(patchResponse, patchRequest)
	if patchResponse.Code != http.StatusOK {
		t.Fatalf("PATCH status = %d: %s", patchResponse.Code, patchResponse.Body.String())
	}
}

func TestProtectedRouteRejectsInsufficientRole(t *testing.T) {
	router := newAuthTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/seller/products", nil)
	req.Header.Set("Authorization", "Bearer buyer-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for insufficient role, got %d", rec.Code)
	}
	assertErrorCode(t, rec, "FORBIDDEN")
}

func TestWebhookRouteRequiresSignatureInsteadOfJWT(t *testing.T) {
	router := newAuthTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/payments/testpay", nil)
	req.Header.Set("Authorization", "Bearer buyer-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing webhook signature, got %d", rec.Code)
	}
	assertErrorCode(t, rec, "UNAUTHORIZED")
}

func TestWebhookRouteAllowsConfiguredSignatureHeader(t *testing.T) {
	router := newAuthTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/payments/testpay", nil)
	req.Header.Set("X-Provider-Signature", "signed")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("expected route bridge response after webhook guard, got %d", rec.Code)
	}
	assertErrorCode(t, rec, "ROUTE_BRIDGE_NOT_CONFIGURED")
}

func TestPaymentRetryRouteProxiesWithInternalPaymentAuth(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/payments/pay_123/retry" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer payment-internal-token-at-least-32-chars" {
			t.Fatalf("authorization=%q", r.Header.Get("Authorization"))
		}
		if r.Header.Get("X-Actor-ID") != "user_buyer" || r.Header.Get("X-Actor-Role") != "buyer" {
			t.Fatalf("actor headers=%v", r.Header)
		}
		if r.Header.Get("Idempotency-Key") != "retry-key-1" {
			t.Fatalf("idempotency key=%q", r.Header.Get("Idempotency-Key"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"payment_id":"pay_retry_1","status":"requires_action"}`)
	}))
	defer upstream.Close()

	path := writeAuthTestContract(t)
	cfg := authTestConfig(path)
	cfg.PaymentHTTPURL = upstream.URL
	cfg.PaymentHTTPTimeout = time.Second
	cfg.PaymentInternalAPIToken = "payment-internal-token-at-least-32-chars"
	repo := repository.NewJSONRouteRepository(path)
	catalog := usecase.NewRouteCatalogService(repo, cfg.APIBasePath)
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	router, err := NewRouterWithOptions(context.Background(), cfg, catalog, logger, nil, RouterOptions{
		TokenVerifier: stubTokenVerifier{},
	})
	if err != nil {
		t.Fatalf("new router: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/payments/pay_123/retry", strings.NewReader(`{"idempotency_key":"retry-key-1"}`))
	req.Header.Set("Authorization", "Bearer buyer-token")
	req.Header.Set("Idempotency-Key", "retry-key-1")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestRouterValidationRejectsInvalidPublicBodyBeforeBridge(t *testing.T) {
	router := newValidationTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", strings.NewReader(`{
		"email":"not-an-email",
		"password":"short",
		"full_name":"Buyer One"
	}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected validation 400, got %d", rec.Code)
	}
	assertErrorCode(t, rec, "VALIDATION_ERROR")
}

func TestRouterValidationRejectsProtectedBodyAfterAuth(t *testing.T) {
	router := newValidationTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/cart/items", strings.NewReader(`{
		"product_id":"prod_12345678",
		"variant_id":"var_12345678",
		"quantity":0
	}`))
	req.Header.Set("Authorization", "Bearer buyer-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected validation 400, got %d", rec.Code)
	}
	assertErrorCode(t, rec, "VALIDATION_ERROR")
}

func TestProtectedRoutesFailClosedWithoutVerifierConfiguration(t *testing.T) {
	path := writeAuthTestContract(t)
	cfg := authTestConfig(path)
	cfg.JWTJWKSURL = ""
	repo := repository.NewJSONRouteRepository(path)
	catalog := usecase.NewRouteCatalogService(repo, cfg.APIBasePath)
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	_, err := NewRouter(context.Background(), cfg, catalog, logger, nil)
	if err == nil {
		t.Fatal("expected protected routes without verifier configuration to fail")
	}
	if !strings.Contains(err.Error(), "JWT_JWKS_URL is required") {
		t.Fatalf("expected missing jwks url error, got %v", err)
	}
}

func newTestRouter(t *testing.T) http.Handler {
	t.Helper()
	return newTestRouterWithHealth(t, nil)
}

func newTestRouterWithHealth(t *testing.T, downstreamHealth DownstreamHealthChecker) http.Handler {
	t.Helper()
	path := filepath.Join(t.TempDir(), "master-api.json")
	body := `{
		"project": "scalable-ecommerce-platform",
		"version": "1.0.0",
		"rest_endpoints": [
			{
				"id": "product.detail",
				"method": "GET",
				"path": "/api/v1/products/{product_id}",
				"service": "product-service",
				"grpc": "ProductService.GetProduct",
				"auth": "public",
				"request_schema": "IdPathRequest",
				"response_schema": "Product"
			}
		]
	}`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write contract: %v", err)
	}

	cfg := config.Config{
		ServiceName:       "api-gateway",
		Environment:       "test",
		HTTPAddress:       ":0",
		APIBasePath:       "/api/v1",
		APIContractPath:   path,
		LogLevel:          "error",
		ReadHeaderTimeout: time.Second,
		ShutdownTimeout:   time.Second,
	}
	repo := repository.NewJSONRouteRepository(path)
	catalog := usecase.NewRouteCatalogService(repo, cfg.APIBasePath)
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	router, err := NewRouter(context.Background(), cfg, catalog, logger, downstreamHealth)
	if err != nil {
		t.Fatalf("new router: %v", err)
	}
	return router
}

func newAuthTestRouter(t *testing.T) http.Handler {
	t.Helper()
	path := writeAuthTestContract(t)
	cfg := authTestConfig(path)
	repo := repository.NewJSONRouteRepository(path)
	catalog := usecase.NewRouteCatalogService(repo, cfg.APIBasePath)
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	router, err := NewRouterWithOptions(context.Background(), cfg, catalog, logger, nil, RouterOptions{
		TokenVerifier: stubTokenVerifier{},
	})
	if err != nil {
		t.Fatalf("new router: %v", err)
	}
	return router
}

func newValidationTestRouter(t *testing.T) http.Handler {
	t.Helper()
	path := writeValidationTestContract(t)
	cfg := authTestConfig(path)
	cfg.Validation = config.RequestValidationConfig{
		Enabled:             true,
		DefaultMaxBodyBytes: 128 * 1024,
		MaxHeaderBytes:      32 * 1024,
		MaxQueryBytes:       8 * 1024,
	}
	repo := repository.NewJSONRouteRepository(path)
	catalog := usecase.NewRouteCatalogService(repo, cfg.APIBasePath)
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	router, err := NewRouterWithOptions(context.Background(), cfg, catalog, logger, nil, RouterOptions{
		TokenVerifier: stubTokenVerifier{},
	})
	if err != nil {
		t.Fatalf("new router: %v", err)
	}
	return router
}

func writeAuthTestContract(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "master-api.json")
	body := `{
		"project": "scalable-ecommerce-platform",
		"version": "1.0.0",
		"rest_endpoints": [
			{
				"id": "user.me_get",
				"method": "GET",
				"path": "/api/v1/me",
				"service": "user-service",
				"grpc": "UserService.GetUser",
				"auth": "buyer",
				"request_schema": "Empty",
				"response_schema": "UserProfile"
			},
			{
				"id": "seller.product_create",
				"method": "POST",
				"path": "/api/v1/seller/products",
				"service": "product-service",
				"grpc": "ProductService.CreateProduct",
				"auth": "seller",
				"request_schema": "ProductInput",
				"response_schema": "Product"
			},
			{
				"id": "notification.preferences_get",
				"method": "GET",
				"path": "/api/v1/me/notification-preferences",
				"service": "notification-service",
				"grpc": "NotificationService.GetNotificationPreference",
				"auth": "buyer",
				"request_schema": "Empty",
				"response_schema": "NotificationPreference"
			},
			{
				"id": "notification.preferences_update",
				"method": "PATCH",
				"path": "/api/v1/me/notification-preferences",
				"service": "notification-service",
				"grpc": "NotificationService.UpdateNotificationPreference",
				"auth": "buyer",
				"request_schema": "NotificationPreferenceInput",
				"response_schema": "NotificationPreference"
			},
			{
				"id": "payment.webhook",
				"method": "POST",
				"path": "/api/v1/webhooks/payments/{provider}",
				"service": "payment-service",
				"grpc": "PaymentService.HandleWebhook",
				"auth": "webhook",
				"request_schema": "PaymentWebhookRequest",
				"response_schema": "SuccessResponse"
			},
			{
				"id": "payment.retry",
				"method": "POST",
				"path": "/api/v1/payments/{payment_id}/retry",
				"service": "payment-service",
				"grpc": "PaymentService.CreatePaymentIntent",
				"auth": "buyer",
				"request_schema": "RetryPaymentRequest",
				"response_schema": "PaymentIntentResponse"
			}
		]
	}`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write contract: %v", err)
	}
	return path
}

func writeValidationTestContract(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "master-api.json")
	body := `{
		"project": "scalable-ecommerce-platform",
		"version": "1.0.0",
		"rest_endpoints": [
			{
				"id": "auth.signup",
				"method": "POST",
				"path": "/api/v1/auth/signup",
				"service": "auth-service",
				"grpc": "AuthService.Signup",
				"auth": "public",
				"request_schema": "SignupRequest",
				"response_schema": "TokenResponse"
			},
			{
				"id": "cart.add_item",
				"method": "POST",
				"path": "/api/v1/cart/items",
				"service": "cart-service",
				"grpc": "CartService.AddItem",
				"auth": "buyer",
				"request_schema": "CartItemInput",
				"response_schema": "Cart"
			}
		],
		"schemas": {
			"SignupRequest": {
				"type": "object",
				"required": ["email", "password", "full_name"],
				"properties": {
					"email": {"type": "string", "format": "email"},
					"password": {"type": "string", "minLength": 8},
					"full_name": {"type": "string"}
				}
			},
			"CartItemInput": {
				"type": "object",
				"required": ["product_id", "variant_id", "quantity"],
				"properties": {
					"product_id": {"type": "string"},
					"variant_id": {"type": "string"},
					"quantity": {"type": "integer", "minimum": 1}
				}
			}
		}
	}`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write contract: %v", err)
	}
	return path
}

func authTestConfig(path string) config.Config {
	return config.Config{
		ServiceName:            "api-gateway",
		Environment:            "test",
		HTTPAddress:            ":0",
		APIBasePath:            "/api/v1",
		APIContractPath:        path,
		LogLevel:               "error",
		ReadHeaderTimeout:      time.Second,
		ShutdownTimeout:        time.Second,
		JWTIssuer:              "ecommerce-auth",
		JWTAudience:            "ecommerce-api",
		JWTAllowedAlgs:         []string{"RS256"},
		JWTJWKSURL:             "http://auth-service/.well-known/jwks.json",
		JWTJWKSCacheTTL:        time.Minute,
		JWTJWKSFetchTimeout:    time.Second,
		JWTClockSkew:           30 * time.Second,
		WebhookSignatureHeader: "X-Provider-Signature",
	}
}

func assertErrorCode(t *testing.T, rec *httptest.ResponseRecorder, code string) {
	t.Helper()
	var envelope ResponseEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if envelope.Error == nil || envelope.Error.Code != code {
		t.Fatalf("expected error code %q, got %+v", code, envelope.Error)
	}
}

type stubDownstreamHealth struct {
	report clients.HealthReport
}

func (s stubDownstreamHealth) Check(context.Context) clients.HealthReport {
	return s.report
}

type stubTokenVerifier struct{}

func (stubTokenVerifier) Verify(_ context.Context, raw string) (gatewayauth.AccessClaims, error) {
	switch raw {
	case "buyer-token":
		return gatewayauth.AccessClaims{
			SessionID: "sess_buyer",
			Roles:     []string{"buyer"},
			TokenType: gatewayauth.AccessTokenType,
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: "user_buyer",
			},
		}, nil
	case "seller-token":
		return gatewayauth.AccessClaims{
			SessionID: "sess_seller",
			Roles:     []string{"seller"},
			SellerID:  "seller_123",
			TokenType: gatewayauth.AccessTokenType,
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: "user_seller",
			},
		}, nil
	case "admin-token":
		return gatewayauth.AccessClaims{
			SessionID: "sess_admin",
			Roles:     []string{"operations_admin"},
			TokenType: gatewayauth.AccessTokenType,
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: "admin_1",
			},
		}, nil
	default:
		return gatewayauth.AccessClaims{}, errors.New("invalid token")
	}
}

type routerUserClient struct {
	getUser       func(context.Context, *userv1.GetUserRequest) (*userv1.UserProfile, error)
	listAddresses func(context.Context, *userv1.ListUserAddressesRequest) (*userv1.AddressListResponse, error)
}

func (c *routerUserClient) GetUser(ctx context.Context, request *userv1.GetUserRequest) (*userv1.UserProfile, error) {
	if c.getUser == nil {
		return nil, errors.New("unexpected GetUser call")
	}
	return c.getUser(ctx, request)
}

func (*routerUserClient) UpdateUserProfile(context.Context, *userv1.UpdateUserProfileRequest) (*userv1.UserProfile, error) {
	return nil, errors.New("unexpected UpdateUserProfile call")
}

func (c *routerUserClient) ListUserAddresses(ctx context.Context, request *userv1.ListUserAddressesRequest) (*userv1.AddressListResponse, error) {
	if c.listAddresses == nil {
		return nil, errors.New("unexpected ListUserAddresses call")
	}
	return c.listAddresses(ctx, request)
}

func (*routerUserClient) CreateAddress(context.Context, *userv1.CreateAddressRequest) (*userv1.Address, error) {
	return nil, errors.New("unexpected CreateAddress call")
}

func (*routerUserClient) UpdateAddress(context.Context, *userv1.UpdateAddressRequest) (*userv1.Address, error) {
	return nil, errors.New("unexpected UpdateAddress call")
}

func (*routerUserClient) DeleteAddress(context.Context, *userv1.DeleteAddressRequest) (*userv1.SuccessResponse, error) {
	return nil, errors.New("unexpected DeleteAddress call")
}

func (*routerUserClient) GetSellerProfile(context.Context, *userv1.GetSellerProfileRequest) (*userv1.SellerProfile, error) {
	return nil, errors.New("unexpected GetSellerProfile call")
}

func (*routerUserClient) UpdateSellerProfile(context.Context, *userv1.UpdateSellerProfileRequest) (*userv1.SellerProfile, error) {
	return nil, errors.New("unexpected UpdateSellerProfile call")
}

type routerNotificationClient struct {
	get    func(context.Context, *notificationv1.GetNotificationPreferenceRequest) (*notificationv1.NotificationPreference, error)
	update func(context.Context, *notificationv1.UpdateNotificationPreferenceRequest) (*notificationv1.NotificationPreference, error)
}

func (c *routerNotificationClient) GetNotificationPreference(ctx context.Context, request *notificationv1.GetNotificationPreferenceRequest) (*notificationv1.NotificationPreference, error) {
	return c.get(ctx, request)
}

func (c *routerNotificationClient) UpdateNotificationPreference(ctx context.Context, request *notificationv1.UpdateNotificationPreferenceRequest) (*notificationv1.NotificationPreference, error) {
	return c.update(ctx, request)
}

func assertNotificationMetadata(t *testing.T, ctx context.Context) {
	t.Helper()
	outgoing, ok := metadata.FromOutgoingContext(ctx)
	if !ok || len(outgoing.Get("x-user-id")) != 1 || outgoing.Get("x-user-id")[0] != "user_buyer" {
		t.Fatalf("outgoing x-user-id = %#v", outgoing.Get("x-user-id"))
	}
	if roles := outgoing.Get("x-roles"); len(roles) != 1 || roles[0] != "buyer" {
		t.Fatalf("outgoing x-roles = %#v", roles)
	}
}

type routerOrderClient struct {
	createOrder       func(context.Context, *orderv1.CreateOrderRequest, ...grpc.CallOption) (*orderv1.CreateOrderResponse, error)
	listSellerOrders  func(context.Context, *orderv1.ListSellerOrdersRequest, ...grpc.CallOption) (*orderv1.ListSellerOrdersResponse, error)
	updateFulfillment func(context.Context, *orderv1.UpdateFulfillmentRequest, ...grpc.CallOption) (*orderv1.UpdateFulfillmentResponse, error)
}

func (c *routerOrderClient) CreateOrder(ctx context.Context, request *orderv1.CreateOrderRequest, opts ...grpc.CallOption) (*orderv1.CreateOrderResponse, error) {
	if c.createOrder == nil {
		return nil, errors.New("unexpected CreateOrder call")
	}
	return c.createOrder(ctx, request, opts...)
}

func (*routerOrderClient) GetOrder(context.Context, *orderv1.GetOrderRequest, ...grpc.CallOption) (*orderv1.GetOrderResponse, error) {
	return nil, errors.New("unexpected GetOrder call")
}

func (*routerOrderClient) ListOrders(context.Context, *orderv1.ListOrdersRequest, ...grpc.CallOption) (*orderv1.ListOrdersResponse, error) {
	return nil, errors.New("unexpected ListOrders call")
}

func (c *routerOrderClient) ListSellerOrders(ctx context.Context, request *orderv1.ListSellerOrdersRequest, opts ...grpc.CallOption) (*orderv1.ListSellerOrdersResponse, error) {
	if c.listSellerOrders == nil {
		return nil, errors.New("unexpected ListSellerOrders call")
	}
	return c.listSellerOrders(ctx, request, opts...)
}

func (*routerOrderClient) CancelOrder(context.Context, *orderv1.CancelOrderRequest, ...grpc.CallOption) (*orderv1.CancelOrderResponse, error) {
	return nil, errors.New("unexpected CancelOrder call")
}

func (c *routerOrderClient) UpdateFulfillment(ctx context.Context, request *orderv1.UpdateFulfillmentRequest, opts ...grpc.CallOption) (*orderv1.UpdateFulfillmentResponse, error) {
	if c.updateFulfillment == nil {
		return nil, errors.New("unexpected UpdateFulfillment call")
	}
	return c.updateFulfillment(ctx, request, opts...)
}
