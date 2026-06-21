package apperrors

import (
	"errors"
	"net/http"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMappings(t *testing.T) {
	tests := []struct {
		code Code
		http int
		grpc codes.Code
	}{
		{CodeValidation, http.StatusBadRequest, codes.InvalidArgument},
		{CodeUnauthorized, http.StatusUnauthorized, codes.Unauthenticated},
		{CodeForbidden, http.StatusForbidden, codes.PermissionDenied},
		{CodeNotFound, http.StatusNotFound, codes.NotFound},
		{CodeConflict, http.StatusConflict, codes.AlreadyExists},
		{CodeRateLimited, http.StatusTooManyRequests, codes.ResourceExhausted},
		{CodeInternal, http.StatusInternalServerError, codes.Internal},
	}
	for _, test := range tests {
		if got := HTTPStatus(test.code); got != test.http {
			t.Errorf("HTTPStatus(%q) = %d, want %d", test.code, got, test.http)
		}
		if got := status.Code(ToGRPC(New(test.code, "safe"))); got != test.grpc {
			t.Errorf("ToGRPC(%q) = %s, want %s", test.code, got, test.grpc)
		}
	}
}

func TestWrapPreservesCause(t *testing.T) {
	cause := errors.New("database details")
	err := Wrap(cause, CodeInternal, "safe")
	if !errors.Is(err, cause) {
		t.Fatal("wrapped error does not preserve cause")
	}
}
