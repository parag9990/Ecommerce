package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	gatewayauth "ecommerce/api-gateway/internal/auth"
	"ecommerce/api-gateway/internal/middleware"
	searchv1 "github.com/parag/ecommerce/backend/shared/gen/go/ecommerce/search/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type SearchHandler struct {
	client  SearchClient
	logger  *slog.Logger
	timeout time.Duration
}

type SearchClient interface {
	SearchProducts(context.Context, *searchv1.SearchRequest, ...grpc.CallOption) (*searchv1.SearchResponse, error)
	Autocomplete(context.Context, *searchv1.AutocompleteRequest, ...grpc.CallOption) (*searchv1.AutocompleteResponse, error)
	CreateSynonym(context.Context, *searchv1.CreateSynonymRequest, ...grpc.CallOption) (*searchv1.SearchSynonym, error)
	UpdateSynonym(context.Context, *searchv1.CreateSynonymRequest, ...grpc.CallOption) (*searchv1.SearchSynonym, error)
	DeleteSynonym(context.Context, *searchv1.CreateSynonymRequest, ...grpc.CallOption) (*searchv1.SearchSynonym, error)
	ListSynonyms(context.Context, *searchv1.ListSynonymsRequest, ...grpc.CallOption) (*searchv1.ListSynonymsResponse, error)
}

type searchHTTPResponse struct {
	Products []*searchv1.Product               `json:"products"`
	Facets   map[string][]*searchv1.FacetValue `json:"facets"`
	Total    int64                             `json:"total"`
}

type autocompleteHTTPResponse struct {
	Suggestions []string `json:"suggestions"`
}

type createSynonymHTTPRequest struct {
	Root     string   `json:"root"`
	Synonyms []string `json:"synonyms"`
	Reason   string   `json:"reason,omitempty"`
	Version  int      `json:"version,omitempty"`
}

type deleteSynonymHTTPRequest struct {
	Reason string `json:"reason,omitempty"`
}

func NewSearchHandler(client SearchClient, logger *slog.Logger) *SearchHandler {
	if client == nil {
		panic("search client is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &SearchHandler{client: client, logger: logger, timeout: 2 * time.Second}
}

func (h *SearchHandler) SearchProducts(w http.ResponseWriter, r *http.Request) {
	request, err := searchGRPCRequest(r)
	if err != nil {
		writeAPIError(w, r, http.StatusBadRequest, "INVALID_SEARCH_REQUEST", err.Error())
		return
	}
	ctx, cancel := h.grpcContext(r)
	defer cancel()
	response, err := h.client.SearchProducts(ctx, request)
	if err != nil {
		h.writeGRPCError(w, r, err)
		return
	}
	facets := make(map[string][]*searchv1.FacetValue, len(response.GetFacets()))
	for _, facet := range response.GetFacets() {
		facets[facet.GetField()] = facet.GetValues()
	}
	writeData(w, r, http.StatusOK, searchHTTPResponse{
		Products: response.GetProducts(),
		Facets:   facets,
		Total:    response.GetTotal(),
	})
}

func (h *SearchHandler) Autocomplete(w http.ResponseWriter, r *http.Request) {
	limit, err := optionalPositiveInt(r.URL.Query().Get("limit"))
	if err != nil {
		writeAPIError(w, r, http.StatusBadRequest, "INVALID_AUTOCOMPLETE_REQUEST", "limit must be a positive integer")
		return
	}
	ctx, cancel := h.grpcContext(r)
	defer cancel()
	response, err := h.client.Autocomplete(ctx, &searchv1.AutocompleteRequest{
		Query: r.URL.Query().Get("q"),
		Limit: int32(limit),
	})
	if err != nil {
		h.writeGRPCError(w, r, err)
		return
	}
	suggestions := make([]string, 0, len(response.GetSuggestions()))
	for _, suggestion := range response.GetSuggestions() {
		if value := strings.TrimSpace(suggestion.GetValue()); value != "" {
			suggestions = append(suggestions, value)
		}
	}
	writeData(w, r, http.StatusOK, autocompleteHTTPResponse{Suggestions: suggestions})
}

func (h *SearchHandler) CreateSynonym(w http.ResponseWriter, r *http.Request) {
	var input createSynonymHTTPRequest
	if err := decodeJSON(w, r, &input); err != nil {
		writeAPIError(w, r, http.StatusBadRequest, "INVALID_SYNONYM", err.Error())
		return
	}
	ctx, cancel := h.grpcContext(r)
	defer cancel()
	if reason := strings.TrimSpace(input.Reason); reason != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "x-audit-reason", reason)
	}
	response, err := h.client.CreateSynonym(ctx, &searchv1.CreateSynonymRequest{
		Root: input.Root, Synonyms: append([]string(nil), input.Synonyms...),
	})
	if err != nil {
		h.writeGRPCError(w, r, err)
		return
	}
	writeData(w, r, http.StatusOK, response)
}

func (h *SearchHandler) UpdateSynonym(w http.ResponseWriter, r *http.Request) {
	synonymID, err := searchSynonymIDFromPath(r.URL.Path)
	if err != nil {
		writeAPIError(w, r, http.StatusBadRequest, "INVALID_SYNONYM", "invalid search synonym")
		return
	}
	var input createSynonymHTTPRequest
	if err := decodeJSON(w, r, &input); err != nil {
		writeAPIError(w, r, http.StatusBadRequest, "INVALID_SYNONYM", err.Error())
		return
	}
	ctx, cancel := h.grpcContext(r)
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, "x-synonym-id", synonymID)
	if reason := strings.TrimSpace(input.Reason); reason != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "x-audit-reason", reason)
	}
	response, err := h.client.UpdateSynonym(ctx, &searchv1.CreateSynonymRequest{
		Root: input.Root, Synonyms: append([]string(nil), input.Synonyms...),
	})
	if err != nil {
		h.writeGRPCError(w, r, err)
		return
	}
	writeData(w, r, http.StatusOK, response)
}

func (h *SearchHandler) DeleteSynonym(w http.ResponseWriter, r *http.Request) {
	synonymID, err := searchSynonymIDFromPath(r.URL.Path)
	if err != nil {
		writeAPIError(w, r, http.StatusBadRequest, "INVALID_SYNONYM", "invalid search synonym")
		return
	}
	var input deleteSynonymHTTPRequest
	if err := decodeJSON(w, r, &input); err != nil {
		writeAPIError(w, r, http.StatusBadRequest, "INVALID_SYNONYM", err.Error())
		return
	}
	ctx, cancel := h.grpcContext(r)
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, "x-synonym-id", synonymID)
	if reason := strings.TrimSpace(input.Reason); reason != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "x-audit-reason", reason)
	}
	if _, err := h.client.DeleteSynonym(ctx, &searchv1.CreateSynonymRequest{Root: synonymID}); err != nil {
		h.writeGRPCError(w, r, err)
		return
	}
	writeData(w, r, http.StatusOK, map[string]bool{"success": true})
}

func (h *SearchHandler) ListSynonyms(w http.ResponseWriter, r *http.Request) {
	page, err := optionalPositiveInt(r.URL.Query().Get("page"))
	if err != nil {
		writeAPIError(w, r, http.StatusBadRequest, "INVALID_SYNONYM_PAGE", "page must be a positive integer")
		return
	}
	pageSize, err := optionalPositiveInt(r.URL.Query().Get("page_size"))
	if err != nil {
		writeAPIError(w, r, http.StatusBadRequest, "INVALID_SYNONYM_PAGE", "page_size must be a positive integer")
		return
	}
	ctx, cancel := h.grpcContext(r)
	defer cancel()
	response, err := h.client.ListSynonyms(ctx, &searchv1.ListSynonymsRequest{Page: int32(page), PageSize: int32(pageSize)})
	if err != nil {
		h.writeGRPCError(w, r, err)
		return
	}
	writeData(w, r, http.StatusOK, response)
}

func searchSynonymIDFromPath(path string) (string, error) {
	value := strings.TrimPrefix(path, "/api/v1/admin/search/synonyms/")
	value = strings.Trim(value, "/")
	if value == "" || strings.Contains(value, "/") {
		return "", strconv.ErrSyntax
	}
	decoded, err := url.PathUnescape(value)
	if err != nil || strings.TrimSpace(decoded) == "" {
		return "", strconv.ErrSyntax
	}
	return strings.TrimSpace(decoded), nil
}

func searchGRPCRequest(r *http.Request) (*searchv1.SearchRequest, error) {
	page, err := optionalPositiveInt(r.URL.Query().Get("page"))
	if err != nil {
		return nil, err
	}
	pageSize, err := optionalPositiveInt(r.URL.Query().Get("page_size"))
	if err != nil {
		return nil, err
	}
	filters := make(map[string]string)
	if rawFilters := strings.TrimSpace(r.URL.Query().Get("filters")); rawFilters != "" {
		var values map[string]any
		if err := json.Unmarshal([]byte(rawFilters), &values); err != nil {
			return nil, err
		}
		for key, value := range values {
			filters[key] = strings.TrimSpace(toFilterString(value))
		}
	}
	for _, key := range []string{"brand", "category_id", "seller_id", "min_price", "max_price", "min_rating", "in_stock"} {
		if value := firstQueryValue(r, key, "filters["+key+"]"); value != "" {
			filters[key] = value
		}
	}
	for _, expression := range r.URL.Query()["filter"] {
		applyFilterExpression(filters, expression)
	}
	return &searchv1.SearchRequest{
		Query: r.URL.Query().Get("q"), Filters: filters, Sort: r.URL.Query().Get("sort"),
		Page: int32(page), PageSize: int32(pageSize),
	}, nil
}

func toFilterString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case bool:
		return strconv.FormatBool(typed)
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	default:
		return ""
	}
}

func applyFilterExpression(filters map[string]string, expression string) {
	expression = strings.TrimSpace(expression)
	for _, item := range []struct{ prefix, key string }{
		{"price:>=", "min_price"}, {"price:<=", "max_price"}, {"rating:>=", "min_rating"},
		{"category_ids:", "category_id"}, {"category_id:", "category_id"}, {"seller_id:", "seller_id"},
		{"brand:", "brand"}, {"in_stock:", "in_stock"},
	} {
		if strings.HasPrefix(expression, item.prefix) {
			filters[item.key] = strings.TrimSpace(strings.TrimPrefix(expression, item.prefix))
			return
		}
	}
}

func firstQueryValue(r *http.Request, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(r.URL.Query().Get(key)); value != "" {
			return value
		}
	}
	return ""
}

func optionalPositiveInt(raw string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return 0, &strconv.NumError{Func: "Atoi", Num: raw, Err: strconv.ErrSyntax}
	}
	return value, nil
}

func (h *SearchHandler) grpcContext(r *http.Request) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	md := metadata.Pairs("x-service-name", "api-gateway", "x-request-id", middleware.RequestIDFromRequest(r))
	for _, header := range []string{"X-Anonymous-ID", "X-Session-ID", "X-Client-Path"} {
		if value := strings.TrimSpace(r.Header.Get(header)); value != "" {
			md.Set(strings.ToLower(header), value)
		}
	}
	if claims, ok := gatewayauth.ClaimsFromContext(r.Context()); ok {
		if userID := strings.TrimSpace(claims.UserID()); userID != "" {
			md.Set("x-user-id", userID)
			md.Set("x-actor-id", userID)
		}
		if len(claims.Roles) > 0 {
			md.Set("x-roles", strings.Join(claims.Roles, ","))
		}
	}
	return metadata.NewOutgoingContext(ctx, md), cancel
}

func (h *SearchHandler) writeGRPCError(w http.ResponseWriter, r *http.Request, err error) {
	code := status.Code(err)
	switch code {
	case codes.InvalidArgument:
		writeAPIError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", status.Convert(err).Message())
	case codes.Unauthenticated:
		writeAPIError(w, r, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "Authentication required")
	case codes.PermissionDenied:
		writeAPIError(w, r, http.StatusForbidden, "PERMISSION_DENIED", "Permission denied")
	case codes.NotFound:
		writeAPIError(w, r, http.StatusNotFound, "NOT_FOUND", status.Convert(err).Message())
	case codes.ResourceExhausted:
		writeAPIError(w, r, http.StatusTooManyRequests, "RATE_LIMITED", "Too many requests")
	case codes.DeadlineExceeded:
		writeAPIError(w, r, http.StatusGatewayTimeout, "UPSTREAM_TIMEOUT", "Search service timed out")
	case codes.Unavailable:
		writeAPIError(w, r, http.StatusServiceUnavailable, "SEARCH_BACKEND_UNAVAILABLE", "Search is temporarily unavailable")
	default:
		h.logger.ErrorContext(r.Context(), "gateway_search_grpc_error", slog.String("code", code.String()))
		writeAPIError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
	}
}
