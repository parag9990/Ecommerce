package grpctransport

import (
	"context"
	"testing"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/search-service/internal/requestctx"
	searchv1 "github.com/parag/ecommerce/backend/shared/gen/go/ecommerce/search/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestSearchProductsMapsRequestResponseAndAnalyticsMetadata(t *testing.T) {
	search := &fakeSearchProducts{execute: func(ctx context.Context, request domain.SearchRequest) (domain.SearchResponse, error) {
		analytics := requestctx.Analytics(ctx)
		if analytics.RequestID != "request_1" || analytics.AnonymousID != "anon_1" || analytics.SessionID != "session_1" {
			t.Fatalf("analytics context = %+v", analytics)
		}
		if request.Query != "phone" || request.Filters[domain.SearchFilterBrand] != "Acme" || request.Page != 2 {
			t.Fatalf("search request = %+v", request)
		}
		return domain.SearchResponse{
			Total:    1,
			Products: []domain.ProductSummary{{ProductID: "product_1", Title: "Phone"}},
			Facets:   map[string][]domain.FacetValue{"brand": {{Value: "Acme", Count: 1}}},
		}, nil
	}}
	handler := newTestHandler(t, search)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		"x-request-id", "request_1", "x-anonymous-id", "anon_1", "x-session-id", "session_1",
	))

	response, err := handler.SearchProducts(ctx, &searchv1.SearchRequest{
		Query: "phone", Filters: map[string]string{domain.SearchFilterBrand: "Acme"}, Page: 2,
	})
	if err != nil {
		t.Fatalf("search products: %v", err)
	}
	if response.GetTotal() != 1 || response.GetProducts()[0].GetProductId() != "product_1" || response.GetFacets()[0].GetField() != "brand" {
		t.Fatalf("search response = %+v", response)
	}
}

func TestSynonymRPCsEnforceAdminIdentityAndRole(t *testing.T) {
	handler := newTestHandler(t, &fakeSearchProducts{})

	_, err := handler.CreateSynonym(context.Background(), &searchv1.CreateSynonymRequest{Root: "phone", Synonyms: []string{"mobile"}})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("anonymous code = %s", status.Code(err))
	}

	nonAdmin := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-actor-id", "user_1", "x-roles", "buyer"))
	_, err = handler.CreateSynonym(nonAdmin, &searchv1.CreateSynonymRequest{Root: "phone", Synonyms: []string{"mobile"}})
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("buyer code = %s", status.Code(err))
	}

	admin := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-actor-id", "admin_1", "x-roles", "catalog_admin", "x-audit-reason", "catalog terminology"))
	response, err := handler.CreateSynonym(admin, &searchv1.CreateSynonymRequest{Root: "phone", Synonyms: []string{"mobile"}})
	if err != nil {
		t.Fatalf("admin create synonym: %v", err)
	}
	if response.GetRoot() != "phone" || response.GetSynonymId() != "syn_1" {
		t.Fatalf("synonym response = %+v", response)
	}
}

func TestUpdateAndDeleteSynonymForwardMetadata(t *testing.T) {
	update := &fakeUpdateSynonym{}
	delete := &fakeDeleteSynonym{}
	handler, err := NewHandler(&fakeSearchProducts{}, fakeAutocomplete{}, fakeCreateSynonym{}, update, delete, fakeListSynonyms{})
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		"x-actor-id", "admin_1",
		"x-roles", "catalog_admin",
		"x-synonym-id", "syn_mobile",
		"x-audit-reason", "catalog terminology",
	))

	if _, err := handler.UpdateSynonym(ctx, &searchv1.CreateSynonymRequest{Root: "cell phone", Synonyms: []string{"smartphone"}}); err != nil {
		t.Fatalf("update synonym: %v", err)
	}
	if update.synonymID != "syn_mobile" || update.input.Root != "cell phone" || update.input.Reason != "catalog terminology" {
		t.Fatalf("update = id:%q input:%#v", update.synonymID, update.input)
	}

	if _, err := handler.DeleteSynonym(ctx, &searchv1.CreateSynonymRequest{Root: "syn_mobile"}); err != nil {
		t.Fatalf("delete synonym: %v", err)
	}
	if delete.synonymID != "syn_mobile" || delete.reason != "catalog terminology" {
		t.Fatalf("delete = id:%q reason:%q", delete.synonymID, delete.reason)
	}
}

func newTestHandler(t *testing.T, search SearchProductsUsecase) *Handler {
	t.Helper()
	handler, err := NewHandler(search, fakeAutocomplete{}, fakeCreateSynonym{}, &fakeUpdateSynonym{}, &fakeDeleteSynonym{}, fakeListSynonyms{})
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}
	return handler
}

type fakeSearchProducts struct {
	execute func(context.Context, domain.SearchRequest) (domain.SearchResponse, error)
}

func (f *fakeSearchProducts) Execute(ctx context.Context, request domain.SearchRequest) (domain.SearchResponse, error) {
	if f.execute != nil {
		return f.execute(ctx, request)
	}
	return domain.SearchResponse{}, nil
}

type fakeAutocomplete struct{}

func (fakeAutocomplete) Execute(context.Context, domain.AutocompleteRequest) (domain.AutocompleteResponse, error) {
	return domain.AutocompleteResponse{}, nil
}

type fakeCreateSynonym struct{}

func (fakeCreateSynonym) Execute(_ context.Context, input domain.SearchSynonymInput) (domain.SearchSynonym, error) {
	return domain.SearchSynonym{ID: "syn_1", Root: input.Root, Synonyms: input.Synonyms}, nil
}

type fakeUpdateSynonym struct {
	synonymID string
	input     domain.SearchSynonymInput
}

func (f *fakeUpdateSynonym) Execute(_ context.Context, synonymID string, input domain.SearchSynonymInput) (domain.SearchSynonym, error) {
	f.synonymID = synonymID
	f.input = input
	return domain.SearchSynonym{ID: "syn_1", Root: input.Root, Synonyms: input.Synonyms}, nil
}

type fakeDeleteSynonym struct {
	synonymID string
	reason    string
}

func (f *fakeDeleteSynonym) Execute(_ context.Context, synonymID string, reason string) (domain.SearchSynonym, error) {
	f.synonymID = synonymID
	f.reason = reason
	return domain.SearchSynonym{ID: synonymID, Root: "mobile", Synonyms: []string{"phone"}}, nil
}

type fakeListSynonyms struct{}

func (fakeListSynonyms) Execute(context.Context, domain.SearchSynonymPageRequest) ([]domain.SearchSynonym, error) {
	return nil, nil
}
