package grpctransport

import (
	"context"
	"errors"
	"sort"
	"strings"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/search-service/internal/requestctx"
	searchv1 "github.com/parag/ecommerce/backend/shared/gen/go/ecommerce/search/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

type SearchProductsUsecase interface {
	Execute(context.Context, domain.SearchRequest) (domain.SearchResponse, error)
}

type AutocompleteUsecase interface {
	Execute(context.Context, domain.AutocompleteRequest) (domain.AutocompleteResponse, error)
}

type CreateSynonymUsecase interface {
	Execute(context.Context, domain.SearchSynonymInput) (domain.SearchSynonym, error)
}

type ListSynonymsUsecase interface {
	Execute(context.Context, domain.SearchSynonymPageRequest) ([]domain.SearchSynonym, error)
}

type Handler struct {
	searchv1.UnimplementedSearchServiceServer
	searchProducts SearchProductsUsecase
	autocomplete   AutocompleteUsecase
	createSynonym  CreateSynonymUsecase
	listSynonyms   ListSynonymsUsecase
}

func NewHandler(searchProducts SearchProductsUsecase, autocomplete AutocompleteUsecase, createSynonym CreateSynonymUsecase, listSynonyms ListSynonymsUsecase) (*Handler, error) {
	if searchProducts == nil {
		return nil, errors.New("search products usecase is required")
	}
	if autocomplete == nil {
		return nil, errors.New("autocomplete usecase is required")
	}
	if createSynonym == nil {
		return nil, errors.New("create synonym usecase is required")
	}
	if listSynonyms == nil {
		return nil, errors.New("list synonyms usecase is required")
	}
	return &Handler{
		searchProducts: searchProducts,
		autocomplete:   autocomplete,
		createSynonym:  createSynonym,
		listSynonyms:   listSynonyms,
	}, nil
}

func (h *Handler) SearchProducts(ctx context.Context, req *searchv1.SearchRequest) (*searchv1.SearchResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	ctx = contextFromMetadata(ctx)
	result, err := h.searchProducts.Execute(ctx, domain.SearchRequest{
		Query:    req.GetQuery(),
		Filters:  cloneStringMap(req.GetFilters()),
		Sort:     req.GetSort(),
		Page:     int(req.GetPage()),
		PageSize: int(req.GetPageSize()),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return searchResponse(result), nil
}

func (h *Handler) Autocomplete(ctx context.Context, req *searchv1.AutocompleteRequest) (*searchv1.AutocompleteResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	ctx = contextFromMetadata(ctx)
	result, err := h.autocomplete.Execute(ctx, domain.AutocompleteRequest{
		Query: req.GetQuery(),
		Limit: int(req.GetLimit()),
	})
	if err != nil {
		return nil, mapError(err)
	}
	suggestions := make([]*searchv1.AutocompleteSuggestion, 0, len(result.Suggestions))
	for _, suggestion := range result.Suggestions {
		suggestions = append(suggestions, &searchv1.AutocompleteSuggestion{
			Value: suggestion,
			Type:  "query",
			Label: suggestion,
		})
	}
	return &searchv1.AutocompleteResponse{Suggestions: suggestions}, nil
}

func (h *Handler) CreateSynonym(ctx context.Context, req *searchv1.CreateSynonymRequest) (*searchv1.SearchSynonym, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	ctx = contextFromMetadata(ctx)
	if err := authorizeAdmin(ctx); err != nil {
		return nil, err
	}
	result, err := h.createSynonym.Execute(ctx, domain.SearchSynonymInput{
		Root:     req.GetRoot(),
		Synonyms: append([]string(nil), req.GetSynonyms()...),
		Reason:   firstMetadataFromContext(ctx, "x-audit-reason"),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return synonymResponse(result), nil
}

func firstMetadataFromContext(ctx context.Context, keys ...string) string {
	md, _ := metadata.FromIncomingContext(ctx)
	return firstMetadata(md, keys...)
}

func (h *Handler) ListSynonyms(ctx context.Context, req *searchv1.ListSynonymsRequest) (*searchv1.ListSynonymsResponse, error) {
	if req == nil {
		req = &searchv1.ListSynonymsRequest{}
	}
	ctx = contextFromMetadata(ctx)
	if err := authorizeAdmin(ctx); err != nil {
		return nil, err
	}
	result, err := h.listSynonyms.Execute(ctx, domain.SearchSynonymPageRequest{
		Page:     int(req.GetPage()),
		PageSize: int(req.GetPageSize()),
	})
	if err != nil {
		return nil, mapError(err)
	}
	items := make([]*searchv1.SearchSynonym, 0, len(result))
	for _, synonym := range result {
		items = append(items, synonymResponse(synonym))
	}
	return &searchv1.ListSynonymsResponse{Synonyms: items}, nil
}

func searchResponse(result domain.SearchResponse) *searchv1.SearchResponse {
	products := make([]*searchv1.Product, 0, len(result.Products))
	for _, product := range result.Products {
		variants := make([]*searchv1.ProductVariant, 0, len(product.Variants))
		for _, variant := range product.Variants {
			var attributes *structpb.Struct
			if len(variant.Attributes) > 0 {
				attributes, _ = structpb.NewStruct(variant.Attributes)
			}
			var price *searchv1.Money
			if variant.Price != nil {
				price = &searchv1.Money{Amount: variant.Price.Amount, Currency: variant.Price.Currency}
			}
			variants = append(variants, &searchv1.ProductVariant{
				Sku:           variant.SKU,
				Attributes:    attributes,
				Price:         price,
				StockQuantity: int32(variant.StockQuantity),
			})
		}
		products = append(products, &searchv1.Product{
			ProductId:   product.ProductID,
			SellerId:    product.SellerID,
			Title:       product.Title,
			Description: product.Description,
			Brand:       product.Brand,
			CategoryId:  product.CategoryID,
			Status:      product.Status,
			Variants:    variants,
		})
	}

	fields := make([]string, 0, len(result.Facets))
	for field := range result.Facets {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	facets := make([]*searchv1.Facet, 0, len(fields))
	for _, field := range fields {
		values := make([]*searchv1.FacetValue, 0, len(result.Facets[field]))
		for _, value := range result.Facets[field] {
			values = append(values, &searchv1.FacetValue{Value: value.Value, Count: int64(value.Count)})
		}
		facets = append(facets, &searchv1.Facet{Field: field, Values: values})
	}
	return &searchv1.SearchResponse{Products: products, Facets: facets, Total: int64(result.Total)}
}

func synonymResponse(value domain.SearchSynonym) *searchv1.SearchSynonym {
	return &searchv1.SearchSynonym{
		SynonymId: value.ID,
		Root:      value.Root,
		Synonyms:  append([]string(nil), value.Synonyms...),
	}
}

func contextFromMetadata(ctx context.Context) context.Context {
	md, _ := metadata.FromIncomingContext(ctx)
	requestID := firstMetadata(md, "x-request-id", "x-correlation-id")
	ctx = requestctx.WithRequestID(ctx, requestID)
	return requestctx.WithAnalyticsContext(ctx, requestctx.AnalyticsContext{
		RequestID:   requestID,
		AnonymousID: firstMetadata(md, "x-anonymous-id"),
		SessionID:   firstMetadata(md, "x-session-id"),
		UserID:      firstMetadata(md, "x-user-id", "x-actor-id"),
		ClientPath:  firstMetadata(md, "x-client-path"),
	})
}

func authorizeAdmin(ctx context.Context) error {
	md, _ := metadata.FromIncomingContext(ctx)
	if firstMetadata(md, "x-user-id", "x-actor-id", "x-admin-id") == "" {
		return status.Error(codes.Unauthenticated, "admin identity is required")
	}
	roles := strings.Split(firstMetadata(md, "x-roles", "x-user-roles", "x-actor-roles"), ",")
	for _, role := range roles {
		switch strings.ToLower(strings.TrimSpace(role)) {
		case "catalog_admin", "superadmin":
			return nil
		}
	}
	return status.Error(codes.PermissionDenied, "search synonym permission is required")
}

func firstMetadata(md metadata.MD, keys ...string) string {
	for _, key := range keys {
		for _, value := range md.Get(key) {
			if value = strings.TrimSpace(value); value != "" {
				return value
			}
		}
	}
	return ""
}

func cloneStringMap(source map[string]string) map[string]string {
	if len(source) == 0 {
		return map[string]string{}
	}
	result := make(map[string]string, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}

func mapError(err error) error {
	switch {
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, "request canceled")
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, "request deadline exceeded")
	case errors.Is(err, domain.ErrInvalidSearchRequest),
		errors.Is(err, domain.ErrUnsupportedSearchSort),
		errors.Is(err, domain.ErrInvalidSearchFilter),
		errors.Is(err, domain.ErrInvalidAutocompleteRequest),
		errors.Is(err, domain.ErrInvalidSynonym):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrUnauthenticated):
		return status.Error(codes.Unauthenticated, "authentication required")
	case errors.Is(err, domain.ErrPermissionDenied):
		return status.Error(codes.PermissionDenied, "permission denied")
	case errors.Is(err, domain.ErrSearchBackendUnavailable),
		errors.Is(err, domain.ErrSearchCollectionUnavailable),
		errors.Is(err, domain.ErrProductHydrationUnavailable),
		errors.Is(err, domain.ErrSchemaUnavailable),
		errors.Is(err, domain.ErrProductIndexUnavailable):
		return status.Error(codes.Unavailable, "search dependency unavailable")
	default:
		return status.Error(codes.Internal, "internal search error")
	}
}
