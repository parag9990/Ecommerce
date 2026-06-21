package httptransport

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/search-service/internal/requestctx"
	"github.com/example/ecommerce-platform/backend/services/search-service/internal/usecase"
)

type SchemaUsecase interface {
	GetProductSchemaContract(ctx context.Context) (usecase.ProductSchemaContract, error)
}

type SearchUsecase interface {
	Execute(ctx context.Context, req domain.SearchRequest) (domain.SearchResponse, error)
}

type AutocompleteUsecase interface {
	Execute(ctx context.Context, req domain.AutocompleteRequest) (domain.AutocompleteResponse, error)
}

type CreateSynonymUsecase interface {
	Execute(ctx context.Context, req domain.SearchSynonymInput) (domain.SearchSynonym, error)
}

type ListSynonymsUsecase interface {
	Execute(ctx context.Context, req domain.SearchSynonymPageRequest) ([]domain.SearchSynonym, error)
}

type ReindexStarter interface {
	Start(ctx context.Context, req domain.ReindexRequest) (domain.ReindexAccepted, error)
}

type ReadinessChecker interface {
	Check(context.Context) error
}

type ReadinessFunc func(context.Context) error

func (f ReadinessFunc) Check(ctx context.Context) error { return f(ctx) }

type Handler struct {
	schemaUsecase       SchemaUsecase
	searchUsecase       SearchUsecase
	autocompleteUsecase AutocompleteUsecase
	createSynonym       CreateSynonymUsecase
	listSynonyms        ListSynonymsUsecase
	reindexStarter      ReindexStarter
	adminAuthorizer     AdminAuthorizer
	adminRateLimiter    AdminRateLimiter
	readiness           ReadinessChecker
	metricsHandler      http.Handler
	logger              *slog.Logger
}

func WithMetricsHandler(metrics http.Handler) HandlerOption {
	return func(h *Handler) error {
		if metrics == nil {
			return errors.New("metrics handler is required")
		}
		h.metricsHandler = metrics
		return nil
	}
}

func WithReadinessChecker(checker ReadinessChecker) HandlerOption {
	return func(h *Handler) error {
		if checker == nil {
			return errors.New("readiness checker is required")
		}
		h.readiness = checker
		return nil
	}
}

type HandlerOption func(*Handler) error

func WithSynonymUsecases(createSynonym CreateSynonymUsecase, listSynonyms ListSynonymsUsecase) HandlerOption {
	return func(h *Handler) error {
		if createSynonym == nil {
			return errors.New("create synonym usecase is required")
		}
		if listSynonyms == nil {
			return errors.New("list synonyms usecase is required")
		}
		h.createSynonym = createSynonym
		h.listSynonyms = listSynonyms
		return nil
	}
}

func WithAdminAuthorizer(authorizer AdminAuthorizer) HandlerOption {
	return func(h *Handler) error {
		if authorizer == nil {
			return errors.New("admin authorizer is required")
		}
		h.adminAuthorizer = authorizer
		return nil
	}
}

func WithAdminRateLimiter(limiter AdminRateLimiter) HandlerOption {
	return func(h *Handler) error {
		if limiter == nil {
			return errors.New("admin rate limiter is required")
		}
		h.adminRateLimiter = limiter
		return nil
	}
}

func WithReindexStarter(starter ReindexStarter) HandlerOption {
	return func(h *Handler) error {
		if starter == nil {
			return errors.New("reindex starter is required")
		}
		h.reindexStarter = starter
		return nil
	}
}

func NewHandler(schemaUsecase SchemaUsecase, searchUsecase SearchUsecase, autocompleteUsecase AutocompleteUsecase, logger *slog.Logger, options ...HandlerOption) (*Handler, error) {
	if schemaUsecase == nil {
		return nil, errors.New("schema usecase is required")
	}
	if searchUsecase == nil {
		return nil, errors.New("search usecase is required")
	}
	if autocompleteUsecase == nil {
		return nil, errors.New("autocomplete usecase is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return (&Handler{
		schemaUsecase:       schemaUsecase,
		searchUsecase:       searchUsecase,
		autocompleteUsecase: autocompleteUsecase,
		adminAuthorizer:     NewHeaderAdminAuthorizer(true),
		adminRateLimiter:    AllowAllAdminRateLimiter{},
		logger:              logger,
	}).applyOptions(options...)
}

func (h *Handler) applyOptions(options ...HandlerOption) (*Handler, error) {
	for _, option := range options {
		if option == nil {
			continue
		}
		if err := option(h); err != nil {
			return nil, err
		}
	}
	return h, nil
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/search", h.handleSearchProducts)
	mux.HandleFunc("/api/v1/search/autocomplete", h.handleAutocomplete)
	mux.HandleFunc("/api/v1/admin/search/reindex", h.handleAdminSearchReindex)
	mux.HandleFunc("/api/v1/admin/search/synonyms", h.handleAdminSearchSynonyms)
	mux.HandleFunc("/internal/v1/search/schema/products", h.handleProductSchema)
	mux.HandleFunc("/readyz", h.handleReady)
	if h.metricsHandler != nil {
		mux.Handle("/metrics", h.metricsHandler)
	}
}

func (h *Handler) handleReady(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	if h.readiness == nil {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := h.readiness.Check(ctx); err != nil {
		h.logger.WarnContext(ctx, "search.readiness.failed", slog.String("error", err.Error()))
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (h *Handler) handleSearchProducts(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	requestID := requestIDFrom(r)
	w.Header().Set("X-Request-ID", requestID)
	ctx := requestctx.WithRequestID(r.Context(), requestID)
	ctx = requestctx.WithAnalyticsContext(ctx, analyticsContextFromRequest(r, requestID))

	req, err := parseSearchRequest(r.URL.Query())
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_SEARCH_REQUEST", "invalid search request")
		return
	}

	resp, err := h.searchUsecase.Execute(ctx, req)
	if err != nil {
		h.writeUsecaseError(w, r.WithContext(ctx), err)
		return
	}

	writeJSON(w, http.StatusOK, searchResponseDTO(resp))
}

func (h *Handler) handleAutocomplete(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	requestID := requestIDFrom(r)
	w.Header().Set("X-Request-ID", requestID)
	ctx := requestctx.WithRequestID(r.Context(), requestID)

	req, err := parseAutocompleteRequest(r.URL.Query())
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_AUTOCOMPLETE_REQUEST", "invalid autocomplete request")
		return
	}

	resp, err := h.autocompleteUsecase.Execute(ctx, req)
	if err != nil {
		h.writeUsecaseError(w, r.WithContext(ctx), err)
		return
	}

	writeJSON(w, http.StatusOK, autocompleteResponseDTO(resp))
}

func (h *Handler) handleAdminSearchSynonyms(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.handleCreateSynonym(w, r)
	case http.MethodGet:
		h.handleListSynonyms(w, r)
	default:
		w.Header().Set("Allow", http.MethodGet+", "+http.MethodPost)
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
	}
}

func (h *Handler) handleAdminSearchReindex(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	requestID := requestIDFrom(r)
	w.Header().Set("X-Request-ID", requestID)
	ctx := requestctx.WithRequestID(r.Context(), requestID)

	actor, err := h.authorizeAdmin(ctx, r, adminPermissionReindexWrite)
	if err != nil {
		h.writeAdminError(w, r.WithContext(ctx), err)
		return
	}
	if !h.adminRateLimiter.Allow(ctx, actor.ID+":search:reindex:write") {
		h.logger.Warn("search.admin.rate_limited",
			slog.String("request_id", requestID),
			slog.String("actor_id", actor.ID),
			slog.String("permission", adminPermissionReindexWrite),
		)
		writeAPIError(w, http.StatusTooManyRequests, "RATE_LIMITED", "too many admin mutation requests")
		return
	}
	if h.reindexStarter == nil {
		h.writeAdminError(w, r.WithContext(ctx), errors.New("reindex starter is not configured"))
		return
	}

	var req reindexRequest
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}

	accepted, err := h.reindexStarter.Start(ctx, domain.ReindexRequest{
		Mode:             req.Mode,
		BatchSize:        req.BatchSize,
		Reason:           req.Reason,
		ActorID:          actor.ID,
		DryRun:           req.DryRun,
		TargetCollection: req.TargetCollection,
	})
	if err != nil {
		h.writeAdminError(w, r.WithContext(ctx), err)
		return
	}

	h.logger.Info("search.admin.audit",
		slog.String("request_id", requestID),
		slog.String("actor_admin_id", actor.ID),
		slog.String("actor_role", strings.Join(actor.Roles, ",")),
		slog.String("action", "search.reindex.requested"),
		slog.String("job_id", accepted.JobID),
		slog.String("mode", req.Mode),
		slog.Bool("dry_run", req.DryRun),
		slog.String("reason", req.Reason),
	)
	writeJSON(w, http.StatusAccepted, reindexAcceptedResponseDTO(accepted))
}

func (h *Handler) handleCreateSynonym(w http.ResponseWriter, r *http.Request) {
	requestID := requestIDFrom(r)
	w.Header().Set("X-Request-ID", requestID)
	ctx := requestctx.WithRequestID(r.Context(), requestID)

	actor, err := h.authorizeAdmin(ctx, r, adminPermissionSynonymsWrite)
	if err != nil {
		h.writeAdminError(w, r.WithContext(ctx), err)
		return
	}
	if !h.adminRateLimiter.Allow(ctx, actor.ID+":search:synonyms:write") {
		h.logger.Warn("search.admin.rate_limited",
			slog.String("request_id", requestID),
			slog.String("actor_id", actor.ID),
			slog.String("permission", adminPermissionSynonymsWrite),
		)
		writeAPIError(w, http.StatusTooManyRequests, "RATE_LIMITED", "too many admin mutation requests")
		return
	}
	if h.createSynonym == nil {
		h.writeAdminError(w, r.WithContext(ctx), errors.New("create synonym usecase is not configured"))
		return
	}

	var req createSynonymRequest
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}

	result, err := h.createSynonym.Execute(ctx, domain.SearchSynonymInput{
		Root:     req.Root,
		Synonyms: req.Synonyms,
		Reason:   req.Reason,
	})
	if err != nil {
		h.writeAdminError(w, r.WithContext(ctx), err)
		return
	}

	h.logSynonymAudit(ctx, actor, "search.synonym.upsert", req.Reason, result)
	writeJSON(w, http.StatusOK, searchSynonymResponseDTO(result))
}

func (h *Handler) handleListSynonyms(w http.ResponseWriter, r *http.Request) {
	requestID := requestIDFrom(r)
	w.Header().Set("X-Request-ID", requestID)
	ctx := requestctx.WithRequestID(r.Context(), requestID)

	actor, err := h.authorizeAdmin(ctx, r, adminPermissionSynonymsRead)
	if err != nil {
		h.writeAdminError(w, r.WithContext(ctx), err)
		return
	}
	if h.listSynonyms == nil {
		h.writeAdminError(w, r.WithContext(ctx), errors.New("list synonyms usecase is not configured"))
		return
	}

	page, err := parseSynonymPageRequest(r.URL.Query())
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_PAGINATION", "invalid pagination request")
		return
	}

	results, err := h.listSynonyms.Execute(ctx, page)
	if err != nil {
		h.writeAdminError(w, r.WithContext(ctx), err)
		return
	}

	h.logger.Info("search.admin.synonym.list",
		slog.String("request_id", requestID),
		slog.String("actor_id", actor.ID),
		slog.String("actor_roles", strings.Join(actor.Roles, ",")),
		slog.Int("synonym_count", len(results)),
	)
	writeJSON(w, http.StatusOK, searchSynonymListResponseDTO(results))
}

func (h *Handler) handleProductSchema(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	contract, err := h.schemaUsecase.GetProductSchemaContract(r.Context())
	if err != nil {
		h.writeUsecaseError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, schemaContractDTO(contract))
}

func (h *Handler) authorizeAdmin(ctx context.Context, r *http.Request, permission string) (AdminActor, error) {
	if h.adminAuthorizer == nil {
		return AdminActor{}, errors.New("admin authorizer is not configured")
	}
	return h.adminAuthorizer.Authorize(ctx, r, permission)
}

func (h *Handler) logSynonymAudit(ctx context.Context, actor AdminActor, action string, reason string, synonym domain.SearchSynonym) {
	h.logger.Info("search.admin.audit",
		slog.String("request_id", requestctx.RequestID(ctx)),
		slog.String("actor_admin_id", actor.ID),
		slog.String("actor_role", strings.Join(actor.Roles, ",")),
		slog.String("action", action),
		slog.String("reason", strings.TrimSpace(reason)),
		slog.String("resource_type", "search_synonym"),
		slog.String("resource_id", synonym.ID),
		slog.String("root", synonym.Root),
		slog.Int("synonym_count", len(synonym.Synonyms)),
	)
}

func (h *Handler) writeUsecaseError(w http.ResponseWriter, r *http.Request, err error) {
	status := http.StatusInternalServerError
	code := "INTERNAL_ERROR"
	message := "internal server error"

	switch {
	case errors.Is(err, domain.ErrInvalidAutocompleteRequest):
		status = http.StatusBadRequest
		code = "INVALID_AUTOCOMPLETE_REQUEST"
		message = "invalid autocomplete request"
	case errors.Is(err, domain.ErrUnsupportedSearchSort):
		status = http.StatusBadRequest
		code = "UNSUPPORTED_SEARCH_SORT"
		message = "unsupported sort"
	case errors.Is(err, domain.ErrInvalidSearchFilter):
		status = http.StatusBadRequest
		code = "INVALID_SEARCH_FILTER"
		message = "invalid filter"
	case errors.Is(err, domain.ErrInvalidSearchRequest):
		status = http.StatusBadRequest
		code = "INVALID_SEARCH_REQUEST"
		message = "invalid search request"
	case errors.Is(err, domain.ErrSearchBackendUnavailable):
		status = http.StatusServiceUnavailable
		code = "SEARCH_BACKEND_UNAVAILABLE"
		message = "search backend unavailable"
	case errors.Is(err, domain.ErrProductHydrationUnavailable):
		status = http.StatusServiceUnavailable
		code = "PRODUCT_SERVICE_UNAVAILABLE"
		message = "product service unavailable"
	case errors.Is(err, domain.ErrInvalidSearchSchema), errors.Is(err, domain.ErrInvalidSearchPolicy), errors.Is(err, domain.ErrInvalidSynonym):
		status = http.StatusInternalServerError
		code = "SCHEMA_CONTRACT_INVALID"
		message = "search schema contract is invalid"
	case errors.Is(err, domain.ErrSchemaUnavailable):
		status = http.StatusServiceUnavailable
		code = "SCHEMA_UNAVAILABLE"
		message = "search schema unavailable"
	case errors.Is(err, context.Canceled):
		status = http.StatusRequestTimeout
		code = "REQUEST_CANCELED"
		message = "request canceled"
	case errors.Is(err, context.DeadlineExceeded):
		status = http.StatusGatewayTimeout
		code = "REQUEST_TIMEOUT"
		message = "request timed out"
	}

	if status >= http.StatusInternalServerError {
		h.logger.Error("search.http.request_failed",
			slog.String("path", r.URL.Path),
			slog.String("method", r.Method),
			slog.Int("status", status),
			slog.String("code", code),
			slog.String("error", err.Error()),
		)
	}
	writeAPIError(w, status, code, message)
}

func (h *Handler) writeAdminError(w http.ResponseWriter, r *http.Request, err error) {
	status := http.StatusInternalServerError
	code := "INTERNAL_ERROR"
	message := "internal server error"

	switch {
	case errors.Is(err, domain.ErrUnauthenticated):
		status = http.StatusUnauthorized
		code = "UNAUTHENTICATED"
		message = "authentication required"
	case errors.Is(err, domain.ErrPermissionDenied):
		status = http.StatusForbidden
		code = "PERMISSION_DENIED"
		message = "admin permission required"
	case errors.Is(err, domain.ErrInvalidSynonym):
		status = http.StatusBadRequest
		code = "INVALID_SYNONYM"
		message = "invalid search synonym"
	case errors.Is(err, domain.ErrInvalidReindexRequest):
		status = http.StatusBadRequest
		code = "INVALID_REINDEX_REQUEST"
		message = "invalid reindex request"
	case errors.Is(err, domain.ErrReindexAlreadyRunning):
		status = http.StatusConflict
		code = "REINDEX_ALREADY_RUNNING"
		message = "search reindex already running"
	case errors.Is(err, domain.ErrSearchCollectionUnavailable), errors.Is(err, domain.ErrSearchBackendUnavailable):
		status = http.StatusServiceUnavailable
		code = "SEARCH_BACKEND_UNAVAILABLE"
		message = "search backend unavailable"
	case errors.Is(err, context.Canceled):
		status = http.StatusRequestTimeout
		code = "REQUEST_CANCELED"
		message = "request canceled"
	case errors.Is(err, context.DeadlineExceeded):
		status = http.StatusGatewayTimeout
		code = "REQUEST_TIMEOUT"
		message = "request timed out"
	}

	if status >= http.StatusInternalServerError || status == http.StatusServiceUnavailable {
		h.logger.Error("search.admin.request_failed",
			slog.String("path", r.URL.Path),
			slog.String("method", r.Method),
			slog.Int("status", status),
			slog.String("code", code),
			slog.String("error", err.Error()),
		)
	}
	writeAPIError(w, status, code, message)
}

func parseSearchRequest(values url.Values) (domain.SearchRequest, error) {
	allowed := map[string]struct{}{
		"q":                           {},
		"sort":                        {},
		"page":                        {},
		"page_size":                   {},
		domain.SearchFilterBrand:      {},
		domain.SearchFilterCategoryID: {},
		domain.SearchFilterSellerID:   {},
		domain.SearchFilterMinPrice:   {},
		domain.SearchFilterMaxPrice:   {},
		domain.SearchFilterMinRating:  {},
		domain.SearchFilterInStock:    {},
	}
	for key := range values {
		if _, ok := allowed[key]; !ok {
			return domain.SearchRequest{}, errors.New("unsupported query parameter")
		}
	}

	page, err := parseIntQuery(values, "page")
	if err != nil {
		return domain.SearchRequest{}, err
	}
	pageSize, err := parseIntQuery(values, "page_size")
	if err != nil {
		return domain.SearchRequest{}, err
	}

	filters := map[string]string{}
	for _, key := range []string{
		domain.SearchFilterBrand,
		domain.SearchFilterCategoryID,
		domain.SearchFilterSellerID,
		domain.SearchFilterMinPrice,
		domain.SearchFilterMaxPrice,
		domain.SearchFilterMinRating,
		domain.SearchFilterInStock,
	} {
		rawValues := values[key]
		if len(rawValues) == 0 {
			continue
		}
		if key == domain.SearchFilterBrand {
			filters[key] = strings.Join(rawValues, ",")
			continue
		}
		if len(rawValues) > 1 {
			return domain.SearchRequest{}, errors.New("repeated filter")
		}
		filters[key] = rawValues[0]
	}

	return domain.SearchRequest{
		Query:    values.Get("q"),
		Filters:  filters,
		Sort:     values.Get("sort"),
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func parseAutocompleteRequest(values url.Values) (domain.AutocompleteRequest, error) {
	allowed := map[string]struct{}{
		"q":     {},
		"limit": {},
	}
	for key := range values {
		if _, ok := allowed[key]; !ok {
			return domain.AutocompleteRequest{}, errors.New("unsupported query parameter")
		}
	}
	if len(values["q"]) > 1 || len(values["limit"]) > 1 {
		return domain.AutocompleteRequest{}, errors.New("repeated query parameter")
	}

	limit, err := parseIntQuery(values, "limit")
	if err != nil {
		return domain.AutocompleteRequest{}, err
	}
	return domain.AutocompleteRequest{
		Query: values.Get("q"),
		Limit: limit,
	}, nil
}

func parseSynonymPageRequest(values url.Values) (domain.SearchSynonymPageRequest, error) {
	allowed := map[string]struct{}{
		"page":      {},
		"page_size": {},
	}
	for key := range values {
		if _, ok := allowed[key]; !ok {
			return domain.SearchSynonymPageRequest{}, errors.New("unsupported query parameter")
		}
	}
	if len(values["page"]) > 1 || len(values["page_size"]) > 1 {
		return domain.SearchSynonymPageRequest{}, errors.New("repeated query parameter")
	}

	page, err := parseIntQuery(values, "page")
	if err != nil {
		return domain.SearchSynonymPageRequest{}, err
	}
	pageSize, err := parseIntQuery(values, "page_size")
	if err != nil {
		return domain.SearchSynonymPageRequest{}, err
	}
	return domain.NormalizeSearchSynonymPageRequest(domain.SearchSynonymPageRequest{
		Page:     page,
		PageSize: pageSize,
	}), nil
}

func parseIntQuery(values url.Values, key string) (int, error) {
	raw := strings.TrimSpace(values.Get(key))
	if raw == "" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0, errors.New("invalid integer query parameter")
	}
	return value, nil
}

func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 64*1024)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return errors.New("request body must contain a single JSON object")
	}
	return nil
}

func requestIDFrom(r *http.Request) string {
	for _, header := range []string{"X-Request-ID", "X-Correlation-ID"} {
		if value := strings.TrimSpace(r.Header.Get(header)); value != "" {
			return value
		}
	}
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "req-unavailable"
	}
	return hex.EncodeToString(b[:])
}

func analyticsContextFromRequest(r *http.Request, requestID string) requestctx.AnalyticsContext {
	return requestctx.AnalyticsContext{
		RequestID:   requestID,
		AnonymousID: r.Header.Get("X-Anonymous-ID"),
		SessionID:   r.Header.Get("X-Session-ID"),
		UserID:      r.Header.Get("X-User-ID"),
		ClientPath:  r.Header.Get("X-Client-Path"),
	}
}
