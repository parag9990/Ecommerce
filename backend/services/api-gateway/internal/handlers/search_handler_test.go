package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	gatewayauth "ecommerce/api-gateway/internal/auth"
	"github.com/golang-jwt/jwt/v5"
	searchv1 "github.com/parag/ecommerce/backend/shared/gen/go/ecommerce/search/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestSearchProductsMapsQueryFiltersAndMetadata(t *testing.T) {
	fake := &fakeSearchClient{
		searchProductsFunc: func(ctx context.Context, request *searchv1.SearchRequest) (*searchv1.SearchResponse, error) {
			if request.GetQuery() != "phone" || request.GetPage() != 2 || request.GetPageSize() != 24 {
				t.Fatalf("unexpected request: %+v", request)
			}
			if request.GetFilters()["min_price"] != "100" || request.GetFilters()["category_id"] != "mobiles" {
				t.Fatalf("unexpected filters: %#v", request.GetFilters())
			}
			md, ok := metadata.FromOutgoingContext(ctx)
			if !ok || firstMetadata(md, "x-anonymous-id") != "anon_123" || firstMetadata(md, "x-session-id") != "session_123" {
				t.Fatalf("unexpected metadata: %#v", md)
			}
			return &searchv1.SearchResponse{Total: 1, Products: []*searchv1.Product{{ProductId: "product_1"}}}, nil
		},
	}
	handler := NewSearchHandler(fake, testLogger())
	request := httptest.NewRequest(http.MethodGet, "/api/v1/search?q=phone&page=2&page_size=24&filter=price:%3E%3D100&filter=category_ids:mobiles", nil)
	request.Header.Set("X-Anonymous-ID", "anon_123")
	request.Header.Set("X-Session-ID", "session_123")
	response := httptest.NewRecorder()

	handler.SearchProducts(response, request)

	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"product_id":"product_1"`) {
		t.Fatalf("status/body = %d %s", response.Code, response.Body.String())
	}
}

func TestCreateSynonymForwardsAuthenticatedAdminMetadata(t *testing.T) {
	fake := &fakeSearchClient{
		createSynonymFunc: func(ctx context.Context, request *searchv1.CreateSynonymRequest) (*searchv1.SearchSynonym, error) {
			md, ok := metadata.FromOutgoingContext(ctx)
			if !ok || firstMetadata(md, "x-actor-id") != "admin_1" || firstMetadata(md, "x-roles") != "catalog_admin" || firstMetadata(md, "x-audit-reason") != "catalog terminology" {
				t.Fatalf("unexpected admin metadata: %#v", md)
			}
			if request.GetRoot() != "mobile" || len(request.GetSynonyms()) != 1 {
				t.Fatalf("unexpected synonym request: %+v", request)
			}
			return &searchv1.SearchSynonym{SynonymId: "syn_1", Root: request.GetRoot(), Synonyms: request.GetSynonyms()}, nil
		},
	}
	handler := NewSearchHandler(fake, testLogger())
	request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/search/synonyms", strings.NewReader(`{"root":"mobile","synonyms":["phone"],"reason":"catalog terminology"}`))
	request = request.WithContext(gatewayauth.WithClaims(request.Context(), gatewayauth.AccessClaims{
		RegisteredClaims: jwt.RegisteredClaims{Subject: "admin_1"},
		Roles:            []string{"catalog_admin"},
	}))
	response := httptest.NewRecorder()

	handler.CreateSynonym(response, request)

	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"synonym_id":"syn_1"`) {
		t.Fatalf("status/body = %d %s", response.Code, response.Body.String())
	}
}

func TestUpdateAndDeleteSynonymForwardPathIDAndReason(t *testing.T) {
	fake := &fakeSearchClient{
		updateSynonymFunc: func(ctx context.Context, request *searchv1.CreateSynonymRequest) (*searchv1.SearchSynonym, error) {
			md, ok := metadata.FromOutgoingContext(ctx)
			if !ok || firstMetadata(md, "x-synonym-id") != "syn_mobile" || firstMetadata(md, "x-audit-reason") != "catalog terminology" {
				t.Fatalf("unexpected update metadata: %#v", md)
			}
			if request.GetRoot() != "cell phone" || len(request.GetSynonyms()) != 1 {
				t.Fatalf("unexpected update request: %+v", request)
			}
			return &searchv1.SearchSynonym{SynonymId: "syn_cell_phone", Root: request.GetRoot(), Synonyms: request.GetSynonyms()}, nil
		},
		deleteSynonymFunc: func(ctx context.Context, request *searchv1.CreateSynonymRequest) (*searchv1.SearchSynonym, error) {
			md, ok := metadata.FromOutgoingContext(ctx)
			if !ok || firstMetadata(md, "x-synonym-id") != "syn_mobile" || firstMetadata(md, "x-audit-reason") != "cleanup" {
				t.Fatalf("unexpected delete metadata: %#v", md)
			}
			if request.GetRoot() != "syn_mobile" {
				t.Fatalf("unexpected delete request: %+v", request)
			}
			return &searchv1.SearchSynonym{SynonymId: "syn_mobile", Root: "mobile"}, nil
		},
	}
	handler := NewSearchHandler(fake, testLogger())

	updateRequest := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/search/synonyms/syn_mobile", strings.NewReader(`{"root":"cell phone","synonyms":["smartphone"],"reason":"catalog terminology"}`))
	updateResponse := httptest.NewRecorder()
	handler.UpdateSynonym(updateResponse, updateRequest)
	if updateResponse.Code != http.StatusOK || !strings.Contains(updateResponse.Body.String(), `"synonym_id":"syn_cell_phone"`) {
		t.Fatalf("update status/body = %d %s", updateResponse.Code, updateResponse.Body.String())
	}

	deleteRequest := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/search/synonyms/syn_mobile", strings.NewReader(`{"reason":"cleanup"}`))
	deleteResponse := httptest.NewRecorder()
	handler.DeleteSynonym(deleteResponse, deleteRequest)
	if deleteResponse.Code != http.StatusOK || !strings.Contains(deleteResponse.Body.String(), `"success":true`) {
		t.Fatalf("delete status/body = %d %s", deleteResponse.Code, deleteResponse.Body.String())
	}
}

func TestSearchProductsMapsUnavailableError(t *testing.T) {
	fake := &fakeSearchClient{searchProductsFunc: func(context.Context, *searchv1.SearchRequest) (*searchv1.SearchResponse, error) {
		return nil, status.Error(codes.Unavailable, "typesense unavailable")
	}}
	handler := NewSearchHandler(fake, testLogger())
	response := httptest.NewRecorder()

	handler.SearchProducts(response, httptest.NewRequest(http.MethodGet, "/api/v1/search?q=phone", nil))

	if response.Code != http.StatusServiceUnavailable || !strings.Contains(response.Body.String(), "SEARCH_BACKEND_UNAVAILABLE") {
		t.Fatalf("status/body = %d %s", response.Code, response.Body.String())
	}
}

func firstMetadata(md metadata.MD, key string) string {
	values := md.Get(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

type fakeSearchClient struct {
	searchProductsFunc func(context.Context, *searchv1.SearchRequest) (*searchv1.SearchResponse, error)
	createSynonymFunc  func(context.Context, *searchv1.CreateSynonymRequest) (*searchv1.SearchSynonym, error)
	updateSynonymFunc  func(context.Context, *searchv1.CreateSynonymRequest) (*searchv1.SearchSynonym, error)
	deleteSynonymFunc  func(context.Context, *searchv1.CreateSynonymRequest) (*searchv1.SearchSynonym, error)
}

func (c *fakeSearchClient) SearchProducts(ctx context.Context, request *searchv1.SearchRequest, _ ...grpc.CallOption) (*searchv1.SearchResponse, error) {
	if c.searchProductsFunc != nil {
		return c.searchProductsFunc(ctx, request)
	}
	return &searchv1.SearchResponse{}, nil
}

func (c *fakeSearchClient) Autocomplete(context.Context, *searchv1.AutocompleteRequest, ...grpc.CallOption) (*searchv1.AutocompleteResponse, error) {
	return &searchv1.AutocompleteResponse{}, nil
}

func (c *fakeSearchClient) CreateSynonym(ctx context.Context, request *searchv1.CreateSynonymRequest, _ ...grpc.CallOption) (*searchv1.SearchSynonym, error) {
	if c.createSynonymFunc != nil {
		return c.createSynonymFunc(ctx, request)
	}
	return &searchv1.SearchSynonym{}, nil
}

func (c *fakeSearchClient) UpdateSynonym(ctx context.Context, request *searchv1.CreateSynonymRequest, _ ...grpc.CallOption) (*searchv1.SearchSynonym, error) {
	if c.updateSynonymFunc != nil {
		return c.updateSynonymFunc(ctx, request)
	}
	return &searchv1.SearchSynonym{}, nil
}

func (c *fakeSearchClient) DeleteSynonym(ctx context.Context, request *searchv1.CreateSynonymRequest, _ ...grpc.CallOption) (*searchv1.SearchSynonym, error) {
	if c.deleteSynonymFunc != nil {
		return c.deleteSynonymFunc(ctx, request)
	}
	return &searchv1.SearchSynonym{}, nil
}

func (c *fakeSearchClient) ListSynonyms(context.Context, *searchv1.ListSynonymsRequest, ...grpc.CallOption) (*searchv1.ListSynonymsResponse, error) {
	return &searchv1.ListSynonymsResponse{}, nil
}
