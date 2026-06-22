package httptransport

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/repository"
	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/usecase"
)

func TestHandlerListTypes(t *testing.T) {
	router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/internal/v1/recommendation/types", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}
	var got typeListResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if len(got.Types) != 4 {
		t.Fatalf("types len = %d, want 4", len(got.Types))
	}
	if got.StrategyIDFormat == "" {
		t.Fatal("strategy id format is empty")
	}
}

func TestHandlerResolveType(t *testing.T) {
	router := newTestRouter(t)

	body := strings.NewReader(`{"context":"cart","cart_product_ids":["prod_1","prod_2"],"limit":4}`)
	req := httptest.NewRequest(http.MethodPost, "/internal/v1/recommendation/resolve-type", body)
	req.Header.Set("X-Request-ID", "req_123")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}
	var got resolveTypeResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if got.Type != "frequently_bought_together" {
		t.Fatalf("type = %s, want frequently_bought_together", got.Type)
	}
	if got.StrategyID != "fbt_v1_order_cooccurrence" {
		t.Fatalf("strategy_id = %s, want fbt_v1_order_cooccurrence", got.StrategyID)
	}
	if got.Request.Limit != 4 {
		t.Fatalf("limit = %d, want 4", got.Request.Limit)
	}
}

func TestHandlerResolveTypeRejectsUnknownFields(t *testing.T) {
	router := newTestRouter(t)

	body := strings.NewReader(`{"context":"home_feed","unexpected":true}`)
	req := httptest.NewRequest(http.MethodPost, "/internal/v1/recommendation/resolve-type", body)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
	assertAPIError(t, rr.Body.Bytes(), "INVALID_REQUEST")
}

func TestHandlerResolveTypeRejectsInvalidContext(t *testing.T) {
	router := newTestRouter(t)

	body := strings.NewReader(`{"context":"unknown"}`)
	req := httptest.NewRequest(http.MethodPost, "/internal/v1/recommendation/resolve-type", body)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
	assertAPIError(t, rr.Body.Bytes(), "VALIDATION_ERROR")
}

func TestHandlerStoragePlan(t *testing.T) {
	router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/internal/v1/recommendation/storage", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}
	var got storagePlanResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if got.DatabaseName != "recommendation_db" {
		t.Fatalf("database_name = %s, want recommendation_db", got.DatabaseName)
	}
	if got.CacheKeyPrefix != "reco:v1" {
		t.Fatalf("cache_key_prefix = %s, want reco:v1", got.CacheKeyPrefix)
	}
	if len(got.MongoCollections) != 9 {
		t.Fatalf("mongo_collections len = %d, want 9", len(got.MongoCollections))
	}
}

func TestHandlerStorageStatus(t *testing.T) {
	router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/internal/v1/recommendation/storage/status", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}
	var got storageStatusResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if !got.Ready {
		t.Fatal("storage status should be ready when no external storage is configured")
	}
	if got.MongoDB.State != "not_configured" {
		t.Fatalf("mongo state = %s, want not_configured", got.MongoDB.State)
	}
}

func TestHandlerReadiness(t *testing.T) {
	router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}
	var got storageStatusResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if !got.Ready {
		t.Fatal("readiness should report ready when external storage is not configured")
	}
}

func newTestRouter(t *testing.T) http.Handler {
	t.Helper()

	repo, err := repository.NewStaticDefinitionRepository()
	if err != nil {
		t.Fatalf("NewStaticDefinitionRepository() error = %v", err)
	}
	service, err := usecase.NewDefinitionService(repo, usecase.DefinitionServiceConfig{
		DefaultLimit:        12,
		MaxLimit:            100,
		MaxIdentifierLength: 128,
	}, slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)))
	if err != nil {
		t.Fatalf("NewDefinitionService() error = %v", err)
	}
	storageService, err := usecase.NewStorageService(nil, nil, usecase.StorageServiceConfig{}, slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)))
	if err != nil {
		t.Fatalf("NewStorageService() error = %v", err)
	}
	handler, err := NewHandler(service, storageService, slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)), 64<<10)
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}
	return NewRouter(handler)
}

func assertAPIError(t *testing.T, body []byte, wantCode string) {
	t.Helper()

	var got errorResponse
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("Unmarshal(errorResponse) error = %v; body=%s", err, string(body))
	}
	if got.Error.Code != wantCode {
		t.Fatalf("error code = %s, want %s", got.Error.Code, wantCode)
	}
}
