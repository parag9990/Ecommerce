package gatewayerrors

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestMapGRPCCodes(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
		wantMsg    string
	}{
		{"invalid argument", status.Error(codes.InvalidArgument, "Invalid cart item"), http.StatusBadRequest, CodeValidation, "Invalid cart item"},
		{"unauthenticated", status.Error(codes.Unauthenticated, "jwt expired"), http.StatusUnauthorized, CodeUnauthorized, "Authentication required"},
		{"permission denied", status.Error(codes.PermissionDenied, "role denied"), http.StatusForbidden, CodeForbidden, "You do not have permission to perform this action"},
		{"not found", status.Error(codes.NotFound, "Product not found"), http.StatusNotFound, CodeNotFound, "Product not found"},
		{"already exists", status.Error(codes.AlreadyExists, "Email already exists"), http.StatusConflict, CodeConflict, "Email already exists"},
		{"aborted", status.Error(codes.Aborted, "Version conflict"), http.StatusConflict, CodeConflict, "Version conflict"},
		{"failed precondition", status.Error(codes.FailedPrecondition, "Order cannot be cancelled after shipment"), http.StatusUnprocessableEntity, CodeFailedPrecondition, "Order cannot be cancelled after shipment"},
		{"out of range", status.Error(codes.OutOfRange, "page size is too large"), http.StatusBadRequest, CodeValidation, "page size is too large"},
		{"resource exhausted", status.Error(codes.ResourceExhausted, "Too many OTP requests"), http.StatusTooManyRequests, CodeRateLimited, "Too many OTP requests"},
		{"deadline exceeded", status.Error(codes.DeadlineExceeded, "service deadline"), http.StatusGatewayTimeout, CodeTimeout, "Request timed out"},
		{"unavailable", status.Error(codes.Unavailable, "dial tcp cart-service: connection refused"), http.StatusServiceUnavailable, CodeServiceUnavailable, "Service temporarily unavailable"},
		{"unimplemented", status.Error(codes.Unimplemented, "missing rpc"), http.StatusNotImplemented, CodeNotImplemented, "Feature is not implemented yet"},
		{"internal", status.Error(codes.Internal, "sql: password=secret"), http.StatusInternalServerError, CodeInternal, "Internal server error"},
		{"unknown", status.Error(codes.Unknown, "panic: stack trace"), http.StatusInternalServerError, CodeInternal, "Internal server error"},
		{"data loss", status.Error(codes.DataLoss, "corrupt internal row"), http.StatusInternalServerError, CodeInternal, "Internal server error"},
		{"cancelled", status.Error(codes.Canceled, "client left"), StatusClientClosedRequest, CodeRequestCancelled, "Request was cancelled"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Map(context.Background(), tt.err)
			if got.Status != tt.wantStatus {
				t.Fatalf("status = %d, want %d", got.Status, tt.wantStatus)
			}
			if got.Code != tt.wantCode {
				t.Fatalf("code = %s, want %s", got.Code, tt.wantCode)
			}
			if got.Message != tt.wantMsg {
				t.Fatalf("message = %q, want %q", got.Message, tt.wantMsg)
			}
		})
	}
}

func TestMapContextErrors(t *testing.T) {
	timeout := Map(context.Background(), context.DeadlineExceeded)
	if timeout.Status != http.StatusGatewayTimeout || timeout.Code != CodeTimeout {
		t.Fatalf("unexpected timeout mapping: %+v", timeout)
	}

	cancelled := Map(context.Background(), context.Canceled)
	if cancelled.Status != StatusClientClosedRequest || cancelled.Code != CodeRequestCancelled {
		t.Fatalf("unexpected cancel mapping: %+v", cancelled)
	}
}

func TestMapLocalError(t *testing.T) {
	err := New(http.StatusTooManyRequests, CodeRateLimited, "Too many requests", map[string]int64{"retry_after_seconds": 60}, map[string]string{"Retry-After": "60"})

	got := Map(context.Background(), err)
	if got.Status != http.StatusTooManyRequests || got.Code != CodeRateLimited {
		t.Fatalf("unexpected local mapping: %+v", got)
	}
	if got.Headers["Retry-After"] != "60" {
		t.Fatalf("expected retry header, got %+v", got.Headers)
	}
}

func TestMapUnknownErrorIsGeneric(t *testing.T) {
	got := Map(context.Background(), errors.New("sql: password=secret failed"))
	if got.Status != http.StatusInternalServerError || got.Code != CodeInternal {
		t.Fatalf("unexpected unknown mapping: %+v", got)
	}
	if got.Message != "Internal server error" {
		t.Fatalf("raw message leaked: %q", got.Message)
	}
}

func TestMapGRPCBadRequestDetails(t *testing.T) {
	st := status.New(codes.InvalidArgument, "Invalid request")
	withDetails, err := st.WithDetails(&errdetails.BadRequest{
		FieldViolations: []*errdetails.BadRequest_FieldViolation{
			{
				Field:       "quantity",
				Reason:      "MIN_VALUE",
				Description: "quantity must be at least 1",
			},
		},
	})
	if err != nil {
		t.Fatalf("attach details: %v", err)
	}

	got := Map(context.Background(), withDetails.Err())
	details, ok := got.Details.([]FieldDetail)
	if !ok {
		t.Fatalf("details type = %T, want []FieldDetail", got.Details)
	}
	if len(details) != 1 || details[0].Field != "quantity" || details[0].Reason != "min_value" {
		t.Fatalf("unexpected details: %+v", details)
	}
	if details[0].Message != "quantity must be at least 1" {
		t.Fatalf("unexpected detail message: %+v", details[0])
	}
}

func TestMapGRPCRetryInfo(t *testing.T) {
	st := status.New(codes.ResourceExhausted, "Too many requests")
	withDetails, err := st.WithDetails(&errdetails.RetryInfo{
		RetryDelay: durationpb.New(90 * time.Second),
	})
	if err != nil {
		t.Fatalf("attach details: %v", err)
	}

	got := Map(context.Background(), withDetails.Err())
	if got.Headers["Retry-After"] != "90" {
		t.Fatalf("expected Retry-After header, got %+v", got.Headers)
	}
	details, ok := got.Details.(map[string]any)
	if !ok {
		t.Fatalf("details type = %T, want map", got.Details)
	}
	if details["retry_after_seconds"] != int64(90) {
		t.Fatalf("expected retry details, got %+v", details)
	}
}

func TestMapGRPCPreconditionAndErrorInfoDetailsAreSanitized(t *testing.T) {
	st := status.New(codes.FailedPrecondition, "Current state blocks this action")
	withDetails, err := st.WithDetails(
		&errdetails.PreconditionFailure{
			Violations: []*errdetails.PreconditionFailure_Violation{
				{
					Type:        "ORDER_STATE",
					Subject:     "orders/order_123",
					Description: "token must never leak",
				},
			},
		},
		&errdetails.ErrorInfo{
			Reason: "STATE_CONFLICT",
			Domain: "order",
			Metadata: map[string]string{
				"retryable": "false",
				"token":     "secret-token",
			},
		},
	)
	if err != nil {
		t.Fatalf("attach details: %v", err)
	}

	got := Map(context.Background(), withDetails.Err())
	details, ok := got.Details.(map[string]any)
	if !ok {
		t.Fatalf("details type = %T, want map", got.Details)
	}
	preconditions := details["preconditions"].([]PreconditionDetail)
	if preconditions[0].Description != "Precondition failed" {
		t.Fatalf("unsafe precondition description leaked: %+v", preconditions[0])
	}
	errorInfos := details["error_info"].([]ErrorInfoDetail)
	if _, exists := errorInfos[0].Metadata["token"]; exists {
		t.Fatalf("unsafe metadata leaked: %+v", errorInfos[0].Metadata)
	}
}
