package http

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"ecommerce/superadmin-service/internal/logging"
)

func TestReadyzWithTypedNilCheckerDoesNotPanic(t *testing.T) {
	var db *sql.DB
	handler := NewRBACHandler(nil, db, logging.NewNop())
	mux := http.NewServeMux()
	handler.Register(mux)

	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	response := httptest.NewRecorder()

	mux.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
}
