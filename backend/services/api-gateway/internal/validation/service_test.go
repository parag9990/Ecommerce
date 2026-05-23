package validation

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ecommerce/api-gateway/internal/domain"
)

func TestServiceAcceptsValidCartItemBody(t *testing.T) {
	service := NewService(testSchemas(), Options{Enabled: true}, nil)
	route := testRoute("cart.add_item", http.MethodPost, "/api/v1/cart/items", "CartItemInput")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cart/items", strings.NewReader(`{
		"product_id":"prod_12345678",
		"variant_id":"var_12345678",
		"quantity":2
	}`))
	req.Header.Set("Content-Type", "application/json")

	validated, err := service.Validate(httptest.NewRecorder(), req, route)
	if err != nil {
		t.Fatalf("expected valid request, got %+v", err)
	}
	payload, ok := BodyPayloadFromContext(validated.Context())
	if !ok {
		t.Fatal("expected validated body payload in context")
	}
	if payload["quantity"] == nil {
		t.Fatalf("expected quantity in payload: %+v", payload)
	}
	if body, readErr := io.ReadAll(validated.Body); readErr != nil || len(body) == 0 {
		t.Fatalf("expected request body to be preserved, body=%q err=%v", string(body), readErr)
	}
}

func TestServiceRejectsUnknownJSONField(t *testing.T) {
	service := NewService(testSchemas(), Options{Enabled: true}, nil)
	route := testRoute("cart.add_item", http.MethodPost, "/api/v1/cart/items", "CartItemInput")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cart/items", strings.NewReader(`{
		"product_id":"prod_12345678",
		"variant_id":"var_12345678",
		"quantity":2,
		"quantitty":2
	}`))
	req.Header.Set("Content-Type", "application/json")

	_, err := service.Validate(httptest.NewRecorder(), req, route)
	assertValidationError(t, err, http.StatusBadRequest, "quantitty", "unknown_field")
}

func TestServiceRejectsUnsupportedContentType(t *testing.T) {
	service := NewService(testSchemas(), Options{Enabled: true}, nil)
	route := testRoute("cart.add_item", http.MethodPost, "/api/v1/cart/items", "CartItemInput")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cart/items", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "text/plain")

	_, err := service.Validate(httptest.NewRecorder(), req, route)
	assertValidationError(t, err, http.StatusUnsupportedMediaType, "Content-Type", "unsupported")
}

func TestServiceRejectsBodyTooLarge(t *testing.T) {
	service := NewService(testSchemas(), Options{Enabled: true, DefaultMaxBodyBytes: 8}, nil)
	route := testRoute("profile.update", http.MethodPatch, "/api/v1/me", "UpdateUserProfileRequest")
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/me", strings.NewReader(`{"full_name":"too large"}`))
	req.Header.Set("Content-Type", "application/json")

	_, err := service.Validate(httptest.NewRecorder(), req, route)
	assertValidationError(t, err, http.StatusRequestEntityTooLarge, "body", "too_large")
}

func TestServiceRejectsQueryPageSizeRange(t *testing.T) {
	service := NewService(testSchemas(), Options{Enabled: true}, nil)
	route := testRoute("order.list", http.MethodGet, "/api/v1/orders", "PaginationRequest")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/orders?page_size=101", nil)

	_, err := service.Validate(httptest.NewRecorder(), req, route)
	assertValidationError(t, err, http.StatusBadRequest, "page_size", "range")
}

func TestServiceRejectsInvalidPathID(t *testing.T) {
	service := NewService(testSchemas(), Options{Enabled: true}, nil)
	route := testRoute("product.detail", http.MethodGet, "/api/v1/products/{product_id}", "IdPathRequest")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/products/bad%20id", nil)
	req.SetPathValue("product_id", "bad id")

	_, err := service.Validate(httptest.NewRecorder(), req, route)
	assertValidationError(t, err, http.StatusBadRequest, "product_id", "invalid")
}

func TestServiceRejectsIdempotencyMismatch(t *testing.T) {
	service := NewService(testSchemas(), Options{Enabled: true}, nil)
	route := testRoute("order.checkout", http.MethodPost, "/api/v1/orders/checkout", "CheckoutRequest")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/orders/checkout", strings.NewReader(`{
		"address_id":"addr_12345678",
		"payment_provider":"stripe",
		"idempotency_key":"body-key-123456789"
	}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "header-key-1234567")

	_, err := service.Validate(httptest.NewRecorder(), req, route)
	assertValidationError(t, err, http.StatusBadRequest, "Idempotency-Key", "mismatch")
}

func TestPasswordValueIsNotEchoedInErrors(t *testing.T) {
	service := NewService(testSchemas(), Options{Enabled: true}, nil)
	route := testRoute("auth.signup", http.MethodPost, "/api/v1/auth/signup", "SignupRequest")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", strings.NewReader(`{
		"email":"buyer@example.com",
		"password":"secret",
		"full_name":"Buyer One"
	}`))
	req.Header.Set("Content-Type", "application/json")

	_, err := service.Validate(httptest.NewRecorder(), req, route)
	if err == nil {
		t.Fatal("expected validation error")
	}
	for _, detail := range err.Details {
		if strings.Contains(detail.Message, "secret") || strings.Contains(detail.Reason, "secret") {
			t.Fatalf("password leaked in validation detail: %+v", detail)
		}
	}
}

func assertValidationError(t *testing.T, err *Error, status int, field string, reason string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected validation error")
	}
	if err.Status != status {
		t.Fatalf("status = %d, want %d; err=%+v", err.Status, status, err)
	}
	for _, detail := range err.Details {
		if detail.Field == field && detail.Reason == reason {
			return
		}
	}
	t.Fatalf("expected detail field=%q reason=%q, got %+v", field, reason, err.Details)
}

func testRoute(id string, method string, path string, schema string) domain.RouteDefinition {
	return domain.RouteDefinition{
		ID:            id,
		Method:        domain.HTTPMethod(method),
		Path:          path,
		Service:       "cart-service",
		GRPCMethod:    "CartService.Test",
		AuthLevel:     domain.AuthPublic,
		RequestSchema: schema,
	}
}

func testSchemas() map[string]domain.Schema {
	minPassword := 8
	minOne := float64(1)
	maxHundred := float64(100)
	return map[string]domain.Schema{
		"SignupRequest": {
			Type:     "object",
			Required: []string{"email", "password", "full_name"},
			Properties: map[string]domain.Schema{
				"email":     {Type: "string", Format: "email"},
				"password":  {Type: "string", MinLength: &minPassword},
				"full_name": {Type: "string"},
			},
		},
		"CartItemInput": {
			Type:     "object",
			Required: []string{"product_id", "variant_id", "quantity"},
			Properties: map[string]domain.Schema{
				"product_id": {Type: "string"},
				"variant_id": {Type: "string"},
				"quantity":   {Type: "integer", Minimum: &minOne},
			},
		},
		"CheckoutRequest": {
			Type:     "object",
			Required: []string{"address_id", "payment_provider", "idempotency_key"},
			Properties: map[string]domain.Schema{
				"address_id":       {Type: "string"},
				"payment_provider": {Type: "string"},
				"idempotency_key":  {Type: "string"},
			},
		},
		"PaginationRequest": {
			Type: "object",
			Properties: map[string]domain.Schema{
				"page":      {Type: "integer", Minimum: &minOne},
				"page_size": {Type: "integer", Minimum: &minOne, Maximum: &maxHundred},
				"cursor":    {Type: "string"},
			},
		},
		"IdPathRequest": {
			Type:       "object",
			Properties: map[string]domain.Schema{"id": {Type: "string"}},
		},
		"UpdateUserProfileRequest": {
			Type:       "object",
			Properties: map[string]domain.Schema{"full_name": {Type: "string"}},
		},
	}
}
