package httptransport

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/search-service/internal/repository"
	"github.com/example/ecommerce-platform/backend/services/search-service/internal/requestctx"
	"github.com/example/ecommerce-platform/backend/services/search-service/internal/schema"
	"github.com/example/ecommerce-platform/backend/services/search-service/internal/usecase"
)

func TestHandleProductSchema(t *testing.T) {
	collection, policy, synonyms := schema.MustProductSchemaContract()
	repo, err := repository.NewStaticSchemaRepository(collection, policy, synonyms)
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	uc, err := usecase.NewSchemaUsecase(repo, slog.Default())
	if err != nil {
		t.Fatalf("new usecase: %v", err)
	}
	handler, err := NewHandler(uc, stubSearchUsecase{}, stubAutocompleteUsecase{}, slog.Default())
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/internal/v1/search/schema/products", nil)
	rec := httptest.NewRecorder()

	NewRouter(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var resp schemaContractResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Collection.Name != "products" {
		t.Fatalf("collection name = %q", resp.Collection.Name)
	}
	if resp.SearchPolicy.DefaultQueryBy != "title,brand,description" {
		t.Fatalf("query_by = %q", resp.SearchPolicy.DefaultQueryBy)
	}
}

func TestHandleSearchProducts(t *testing.T) {
	searchUC := &capturingSearchUsecase{
		resp: domain.SearchResponse{
			Products: []domain.ProductSummary{
				{
					ProductID:  "prod_1",
					SellerID:   "seller_1",
					Title:      "Nike Running Shoes",
					Brand:      "Nike",
					CategoryID: "cat_shoes",
					Status:     "published",
				},
			},
			Facets: map[string][]domain.FacetValue{
				domain.ProductFieldBrand: {{Value: "Nike", Count: 1}},
			},
			Total: 1,
		},
	}
	handler, err := NewHandler(stubSchemaUsecase{}, searchUC, stubAutocompleteUsecase{}, slog.Default())
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/search?q=running+shoes&brand=Nike&brand=Puma&category_id=cat_shoes&page=2&page_size=10&sort=price_asc", nil)
	req.Header.Set("X-Request-ID", "req_123")
	req.Header.Set("X-Anonymous-ID", "anon_123")
	req.Header.Set("X-Session-ID", "sess_123")
	req.Header.Set("X-User-ID", "user_123")
	req.Header.Set("X-Client-Path", "/search?ignored=true")
	rec := httptest.NewRecorder()

	NewRouter(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("X-Request-ID") != "req_123" {
		t.Fatalf("request id header = %q", rec.Header().Get("X-Request-ID"))
	}
	if searchUC.req.Query != "running shoes" {
		t.Fatalf("query = %q", searchUC.req.Query)
	}
	if searchUC.req.Filters[domain.SearchFilterBrand] != "Nike,Puma" {
		t.Fatalf("brand filter = %q", searchUC.req.Filters[domain.SearchFilterBrand])
	}
	analytics := requestctx.Analytics(searchUC.ctx)
	if analytics.RequestID != "req_123" || analytics.AnonymousID != "anon_123" || analytics.SessionID != "sess_123" || analytics.UserID != "user_123" {
		t.Fatalf("analytics context = %#v", analytics)
	}
	if analytics.ClientPath != "/search" {
		t.Fatalf("client path = %q", analytics.ClientPath)
	}

	var resp searchProductsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Total != 1 || len(resp.Products) != 1 || resp.Products[0].ProductID != "prod_1" {
		t.Fatalf("unexpected response: %#v", resp)
	}
}

func TestHandleSearchProductsRejectsInvalidQuery(t *testing.T) {
	handler, err := NewHandler(stubSchemaUsecase{}, stubSearchUsecase{}, stubAutocompleteUsecase{}, slog.Default())
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/search?page=abc", nil)
	rec := httptest.NewRecorder()

	NewRouter(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestHandleAutocomplete(t *testing.T) {
	autocompleteUC := &capturingAutocompleteUsecase{
		resp: domain.AutocompleteResponse{Suggestions: []string{"shoes", "shoe rack"}},
	}
	handler, err := NewHandler(stubSchemaUsecase{}, stubSearchUsecase{}, autocompleteUC, slog.Default())
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/search/autocomplete?q=sho&limit=2", nil)
	req.Header.Set("X-Request-ID", "req_auto")
	rec := httptest.NewRecorder()

	NewRouter(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("X-Request-ID") != "req_auto" {
		t.Fatalf("request id header = %q", rec.Header().Get("X-Request-ID"))
	}
	if autocompleteUC.req.Query != "sho" || autocompleteUC.req.Limit != 2 {
		t.Fatalf("request = %#v", autocompleteUC.req)
	}

	var resp autocompleteResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Suggestions) != 2 || resp.Suggestions[0] != "shoes" {
		t.Fatalf("unexpected response: %#v", resp)
	}
}

func TestHandleCreateSynonym(t *testing.T) {
	createUC := &capturingCreateSynonymUsecase{
		resp: domain.SearchSynonym{ID: "syn_mobile", Root: "mobile", Synonyms: []string{"phone", "smartphone"}},
	}
	handler, err := NewHandler(
		stubSchemaUsecase{},
		stubSearchUsecase{},
		stubAutocompleteUsecase{},
		slog.Default(),
		WithSynonymUsecases(createUC, stubListSynonymsUsecase{}),
	)
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}

	body := `{"root":" Mobile ","synonyms":[" Phone ","SMARTPHONE"]}`
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/admin/search/synonyms", strings.NewReader(body))
	req.Header.Set("X-Request-ID", "req_syn")
	req.Header.Set("X-User-ID", "admin_1")
	req.Header.Set("X-Roles", "catalog_admin")
	rec := httptest.NewRecorder()

	NewRouter(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if createUC.req.Root != " Mobile " || len(createUC.req.Synonyms) != 2 {
		t.Fatalf("request = %#v", createUC.req)
	}
	var resp searchSynonymResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.SynonymID != "syn_mobile" || resp.Root != "mobile" {
		t.Fatalf("response = %#v", resp)
	}
}

func TestHandleCreateSynonymRequiresAdminRole(t *testing.T) {
	handler, err := NewHandler(
		stubSchemaUsecase{},
		stubSearchUsecase{},
		stubAutocompleteUsecase{},
		slog.Default(),
		WithSynonymUsecases(stubCreateSynonymUsecase{}, stubListSynonymsUsecase{}),
	)
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/admin/search/synonyms", strings.NewReader(`{"root":"mobile","synonyms":["phone"]}`))
	req.Header.Set("X-User-ID", "buyer_1")
	req.Header.Set("X-Roles", "buyer")
	rec := httptest.NewRecorder()

	NewRouter(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestHandleListSynonyms(t *testing.T) {
	listUC := &capturingListSynonymsUsecase{
		resp: []domain.SearchSynonym{{ID: "syn_mobile", Root: "mobile", Synonyms: []string{"phone"}}},
	}
	handler, err := NewHandler(
		stubSchemaUsecase{},
		stubSearchUsecase{},
		stubAutocompleteUsecase{},
		slog.Default(),
		WithSynonymUsecases(stubCreateSynonymUsecase{}, listUC),
	)
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/admin/search/synonyms?page=2&page_size=10", nil)
	req.Header.Set("X-User-ID", "admin_1")
	req.Header.Set("X-Roles", "superadmin")
	rec := httptest.NewRecorder()

	NewRouter(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if listUC.req.Page != 2 || listUC.req.PageSize != 10 {
		t.Fatalf("page request = %#v", listUC.req)
	}
	var resp searchSynonymListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Synonyms) != 1 || resp.Synonyms[0].SynonymID != "syn_mobile" {
		t.Fatalf("response = %#v", resp)
	}
}

func TestHandleAdminSearchReindex(t *testing.T) {
	starter := &capturingReindexStarter{accepted: domain.ReindexAccepted{JobID: "search_reindex_20260524_120000", Status: "accepted"}}
	handler, err := NewHandler(
		stubSchemaUsecase{},
		stubSearchUsecase{},
		stubAutocompleteUsecase{},
		slog.Default(),
		WithReindexStarter(starter),
	)
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}

	body := `{"mode":"alias","batch_size":500,"reason":"schema_change","dry_run":true}`
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/admin/search/reindex", strings.NewReader(body))
	req.Header.Set("X-Request-ID", "req_reindex")
	req.Header.Set("X-User-ID", "admin_1")
	req.Header.Set("X-Roles", "superadmin")
	rec := httptest.NewRecorder()

	NewRouter(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if starter.req.Mode != "alias" || starter.req.BatchSize != 500 || starter.req.Reason != "schema_change" || starter.req.ActorID != "admin_1" || !starter.req.DryRun {
		t.Fatalf("request = %#v", starter.req)
	}
	var resp reindexAcceptedResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.JobID != "search_reindex_20260524_120000" || resp.Status != "accepted" {
		t.Fatalf("response = %#v", resp)
	}
}

type stubSchemaUsecase struct{}

func (stubSchemaUsecase) GetProductSchemaContract(context.Context) (usecase.ProductSchemaContract, error) {
	collection, policy, synonyms := schema.MustProductSchemaContract()
	return usecase.ProductSchemaContract{
		Version:             domain.ProductSchemaContractVersion,
		Collection:          collection,
		TypesenseDefinition: map[string]any{"name": collection.Name},
		SearchPolicy:        policy,
		SynonymModel:        synonyms,
	}, nil
}

type stubSearchUsecase struct{}

func (stubSearchUsecase) Execute(context.Context, domain.SearchRequest) (domain.SearchResponse, error) {
	return domain.SearchResponse{Products: []domain.ProductSummary{}, Facets: map[string][]domain.FacetValue{}, Total: 0}, nil
}

type stubAutocompleteUsecase struct{}

func (stubAutocompleteUsecase) Execute(context.Context, domain.AutocompleteRequest) (domain.AutocompleteResponse, error) {
	return domain.AutocompleteResponse{Suggestions: []string{}}, nil
}

type stubCreateSynonymUsecase struct{}

func (stubCreateSynonymUsecase) Execute(context.Context, domain.SearchSynonymInput) (domain.SearchSynonym, error) {
	return domain.SearchSynonym{ID: "syn_mobile", Root: "mobile", Synonyms: []string{"phone"}}, nil
}

type stubListSynonymsUsecase struct{}

func (stubListSynonymsUsecase) Execute(context.Context, domain.SearchSynonymPageRequest) ([]domain.SearchSynonym, error) {
	return []domain.SearchSynonym{}, nil
}

type capturingSearchUsecase struct {
	ctx  context.Context
	req  domain.SearchRequest
	resp domain.SearchResponse
}

func (u *capturingSearchUsecase) Execute(ctx context.Context, req domain.SearchRequest) (domain.SearchResponse, error) {
	u.ctx = ctx
	u.req = req
	return u.resp, nil
}

type capturingAutocompleteUsecase struct {
	req  domain.AutocompleteRequest
	resp domain.AutocompleteResponse
}

func (u *capturingAutocompleteUsecase) Execute(_ context.Context, req domain.AutocompleteRequest) (domain.AutocompleteResponse, error) {
	u.req = req
	return u.resp, nil
}

type capturingCreateSynonymUsecase struct {
	req  domain.SearchSynonymInput
	resp domain.SearchSynonym
}

func (u *capturingCreateSynonymUsecase) Execute(_ context.Context, req domain.SearchSynonymInput) (domain.SearchSynonym, error) {
	u.req = req
	return u.resp, nil
}

type capturingListSynonymsUsecase struct {
	req  domain.SearchSynonymPageRequest
	resp []domain.SearchSynonym
}

func (u *capturingListSynonymsUsecase) Execute(_ context.Context, req domain.SearchSynonymPageRequest) ([]domain.SearchSynonym, error) {
	u.req = req
	return u.resp, nil
}

type capturingReindexStarter struct {
	req      domain.ReindexRequest
	accepted domain.ReindexAccepted
	err      error
}

func (s *capturingReindexStarter) Start(_ context.Context, req domain.ReindexRequest) (domain.ReindexAccepted, error) {
	s.req = req
	if s.err != nil {
		return domain.ReindexAccepted{}, s.err
	}
	return s.accepted, nil
}
