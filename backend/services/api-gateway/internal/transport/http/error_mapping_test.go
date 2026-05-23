package httptransport

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestWriteMappedErrorEnvelopeFromGRPC(t *testing.T) {
	handler := RequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeMappedError(w, r, status.Error(codes.NotFound, "Product not found"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/products/prod_missing", nil)
	req.Header.Set("X-Request-Id", "req_error_map")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	var envelope ResponseEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if envelope.Data != nil {
		t.Fatalf("error response data must be null, got %+v", envelope.Data)
	}
	if envelope.RequestID != "req_error_map" {
		t.Fatalf("request_id = %q, want req_error_map", envelope.RequestID)
	}
	if envelope.Error == nil || envelope.Error.Code != "NOT_FOUND" || envelope.Error.Message != "Product not found" {
		t.Fatalf("unexpected error envelope: %+v", envelope.Error)
	}
}

func TestWriteMappedErrorAppliesRetryInfo(t *testing.T) {
	st := status.New(codes.ResourceExhausted, "Too many requests")
	withDetails, err := st.WithDetails(&errdetails.RetryInfo{
		RetryDelay: durationpb.New(45 * time.Second),
	})
	if err != nil {
		t.Fatalf("attach details: %v", err)
	}
	handler := RequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeMappedError(w, r, withDetails.Err())
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/otp/send", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusTooManyRequests)
	}
	if rec.Header().Get("Retry-After") != "45" {
		t.Fatalf("Retry-After = %q, want 45", rec.Header().Get("Retry-After"))
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	errorBody := body["error"].(map[string]any)
	details := errorBody["details"].(map[string]any)
	if details["retry_after_seconds"].(float64) != 45 {
		t.Fatalf("unexpected retry details: %+v", details)
	}
}

func TestWriteMappedErrorSanitizesInternalGRPCMessage(t *testing.T) {
	handler := RequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeMappedError(w, r, status.Error(codes.Internal, "sql: password=secret"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/orders", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	if body := rec.Body.String(); body == "" || containsAny(body, "sql:", "password", "secret") {
		t.Fatalf("internal detail leaked in response: %s", body)
	}
}

func containsAny(value string, fragments ...string) bool {
	for _, fragment := range fragments {
		if fragment != "" && strings.Contains(value, fragment) {
			return true
		}
	}
	return false
}
