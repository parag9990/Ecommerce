package httptransport

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/usecase"
)

type TypeDefinitionUsecase interface {
	ListTypes(ctx context.Context) ([]domain.TypeDefinition, error)
	ListContexts(ctx context.Context) ([]domain.ContextDefinition, error)
	Resolve(ctx context.Context, input usecase.ResolveInput) (usecase.Resolution, error)
}

type StorageUsecase interface {
	StoragePlan(ctx context.Context) (domain.StoragePlan, error)
	Status(ctx context.Context) (usecase.StorageStatus, error)
}

type Handler struct {
	usecase        TypeDefinitionUsecase
	storageUsecase StorageUsecase
	logger         *slog.Logger
	maxBodyBytes   int64
}

func NewHandler(usecase TypeDefinitionUsecase, storageUsecase StorageUsecase, logger *slog.Logger, maxBodyBytes int64) (*Handler, error) {
	if usecase == nil {
		return nil, errors.New("type definition usecase is required")
	}
	if storageUsecase == nil {
		return nil, errors.New("storage usecase is required")
	}
	if maxBodyBytes <= 0 {
		return nil, errors.New("max body bytes must be greater than zero")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{
		usecase:        usecase,
		storageUsecase: storageUsecase,
		logger:         logger,
		maxBodyBytes:   maxBodyBytes,
	}, nil
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/readyz", h.handleReadiness)
	mux.HandleFunc("/internal/v1/recommendation/types", h.handleListTypes)
	mux.HandleFunc("/internal/v1/recommendation/contexts", h.handleListContexts)
	mux.HandleFunc("/internal/v1/recommendation/resolve-type", h.handleResolveType)
	mux.HandleFunc("/internal/v1/recommendation/storage", h.handleStoragePlan)
	mux.HandleFunc("/internal/v1/recommendation/storage/status", h.handleStorageStatus)
}

func (h *Handler) handleReadiness(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	status, err := h.storageUsecase.Status(r.Context())
	if err != nil {
		h.writeUsecaseError(w, r, err)
		return
	}
	statusCode := http.StatusOK
	if !status.Ready() {
		statusCode = http.StatusServiceUnavailable
	}
	writeJSON(w, statusCode, storageStatusFromUsecase(status))
}

func (h *Handler) handleListTypes(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	types, err := h.usecase.ListTypes(r.Context())
	if err != nil {
		h.writeUsecaseError(w, r, err)
		return
	}
	response := typeListResponse{
		Types:            make([]typeDefinitionResponse, 0, len(types)),
		StrategyIDFormat: domain.StrategyIDPattern,
	}
	for _, definition := range types {
		response.Types = append(response.Types, typeDefinitionFromDomain(definition))
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) handleListContexts(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	contexts, err := h.usecase.ListContexts(r.Context())
	if err != nil {
		h.writeUsecaseError(w, r, err)
		return
	}
	response := contextListResponse{
		Contexts: make([]contextDefinitionResponse, 0, len(contexts)),
	}
	for _, definition := range contexts {
		response.Contexts = append(response.Contexts, contextDefinitionFromDomain(definition))
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) handleResolveType(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var req resolveTypeRequest
	if err := h.decodeJSON(w, r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	resolution, err := h.usecase.Resolve(r.Context(), usecase.ResolveInput{
		RequestID:      requestID(r),
		UserID:         req.UserID,
		AnonymousID:    req.AnonymousID,
		SessionID:      req.SessionID,
		ProductID:      req.ProductID,
		CategoryID:     req.CategoryID,
		SellerID:       req.SellerID,
		CartProductIDs: req.CartProductIDs,
		Context:        domain.RecommendationContext(req.Context),
		Limit:          req.Limit,
		Type:           domain.RecommendationType(req.Type),
	})
	if err != nil {
		h.writeUsecaseError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, resolveTypeResponse{
		Request:          recommendationRequestFromDomain(resolution.Request),
		Type:             string(resolution.Type),
		StrategyID:       string(resolution.StrategyID),
		StrategyIDFormat: resolution.StrategyFormat,
		Context:          contextDefinitionFromDomain(resolution.Context),
		Definition:       typeDefinitionFromDomain(resolution.Definition),
		Fallbacks:        fallbackStepsFromDomain(resolution.Fallbacks),
	})
}

func (h *Handler) handleStoragePlan(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	plan, err := h.storageUsecase.StoragePlan(r.Context())
	if err != nil {
		h.writeUsecaseError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, storagePlanFromDomain(plan))
}

func (h *Handler) handleStorageStatus(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	status, err := h.storageUsecase.Status(r.Context())
	if err != nil {
		h.writeUsecaseError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, storageStatusFromUsecase(status))
}

func (h *Handler) decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, h.maxBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

func (h *Handler) writeUsecaseError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidRecommendationRequest),
		errors.Is(err, domain.ErrInvalidLimit),
		errors.Is(err, domain.ErrUnsupportedContext),
		errors.Is(err, domain.ErrUnsupportedRecommendationType),
		errors.Is(err, domain.ErrUnsupportedStrategy),
		errors.Is(err, domain.ErrInvalidCacheKey),
		errors.Is(err, domain.ErrInvalidRecommendationSet):
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	case errors.Is(err, domain.ErrDefinitionNotFound):
		writeAPIError(w, http.StatusNotFound, "NOT_FOUND", "Recommendation definition not found")
	case errors.Is(err, domain.ErrRecommendationStorage):
		writeAPIError(w, http.StatusServiceUnavailable, "STORAGE_UNAVAILABLE", "Recommendation storage is unavailable")
	case errors.Is(err, context.Canceled):
		writeAPIError(w, http.StatusRequestTimeout, "REQUEST_CANCELLED", "Request cancelled")
	case errors.Is(err, context.DeadlineExceeded):
		writeAPIError(w, http.StatusGatewayTimeout, "REQUEST_TIMEOUT", "Request timed out")
	default:
		h.logger.ErrorContext(r.Context(), "recommendation.http_error",
			slog.String("request_id", requestID(r)),
			slog.String("error", err.Error()),
		)
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
	}
}

func requireMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method == method {
		return true
	}
	w.Header().Set("Allow", method)
	writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
	return false
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeAPIError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, errorResponse{
		Error: apiError{
			Code:    code,
			Message: message,
		},
	})
}

func requestID(r *http.Request) string {
	if value := strings.TrimSpace(r.Header.Get("X-Request-ID")); value != "" {
		return value
	}
	if value := strings.TrimSpace(r.Header.Get("Traceparent")); value != "" {
		return value
	}
	return ""
}
