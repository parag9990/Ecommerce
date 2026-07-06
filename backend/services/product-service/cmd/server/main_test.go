package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"product-service/internal/domain"
	"product-service/internal/usecase"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestAuthorizedInternalServiceFailsClosed(t *testing.T) {
	request := httptest.NewRequest("GET", "/internal/v1/products/search-export", nil)
	request.Header.Set("X-Service-Token", "service-token")

	if authorizedInternalService(request, "") {
		t.Fatal("empty configured token must fail closed")
	}
	if authorizedInternalService(request, "different-token") {
		t.Fatal("mismatched token must be rejected")
	}
	if !authorizedInternalService(request, "service-token") {
		t.Fatal("matching token must be accepted")
	}
}

func TestInternalServiceAuthProtectsEveryInternalRoute(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	handler := internalServiceAuth(next, "service-token")

	for _, test := range []struct {
		name       string
		path       string
		token      string
		wantStatus int
	}{
		{name: "missing internal token", path: "/internal/v1/products/batch", wantStatus: http.StatusUnauthorized},
		{name: "wrong internal token", path: "/internal/v1/inventory/reservations", token: "wrong", wantStatus: http.StatusUnauthorized},
		{name: "valid internal token", path: "/internal/v1/products/batch", token: "service-token", wantStatus: http.StatusNoContent},
		{name: "public route", path: "/api/v1/products", wantStatus: http.StatusNoContent},
	} {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			if test.token != "" {
				request.Header.Set("X-Service-Token", test.token)
			}
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, test.wantStatus)
			}
		})
	}
}

func TestInternalGRPCAuth(t *testing.T) {
	interceptor := internalGRPCAuth("service-token")
	handler := func(context.Context, any) (any, error) { return "ok", nil }
	internalInfo := &grpc.UnaryServerInfo{FullMethod: "/ecommerce.product.v1.ProductService/ReserveInventory"}
	publicInfo := &grpc.UnaryServerInfo{FullMethod: "/ecommerce.product.v1.ProductService/ListProducts"}

	if _, err := interceptor(context.Background(), nil, internalInfo, handler); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("missing token error = %v, want unauthenticated", err)
	}
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-service-token", "service-token"))
	if _, err := interceptor(ctx, nil, internalInfo, handler); err != nil {
		t.Fatalf("valid token rejected: %v", err)
	}
	if _, err := interceptor(context.Background(), nil, publicInfo, handler); err != nil {
		t.Fatalf("public method rejected: %v", err)
	}
}

func TestWriteServiceErrorIncludesValidationDetails(t *testing.T) {
	response := httptest.NewRecorder()
	writeServiceError(response, &usecase.ServiceError{
		Kind:    usecase.ErrorKindInvalidArgument,
		Code:    usecase.ErrorCodeValidation,
		Message: "validation failed",
		Report: domain.ValidationReport{Issues: []domain.ValidationIssue{{
			Code:     domain.CodeUnknownAttributeKey,
			Field:    "attributes.neck_type",
			Message:  "attribute is not allowed by category schema",
			Severity: domain.IssueSeverityError,
		}}},
	})

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}

	var body struct {
		Error struct {
			Code    string                   `json:"code"`
			Details []domain.ValidationIssue `json:"details"`
		} `json:"error"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Error.Code != usecase.ErrorCodeValidation {
		t.Fatalf("code = %s, want %s", body.Error.Code, usecase.ErrorCodeValidation)
	}
	if len(body.Error.Details) != 1 || body.Error.Details[0].Field != "attributes.neck_type" {
		t.Fatalf("details = %+v, want validation issue", body.Error.Details)
	}
}
