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
	"strings"

	"github.com/example/ecommerce-platform/backend/services/session-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/session-service/internal/usecase"
)

const maxRequestBodyBytes = 1 << 20

type PrivacyUsecase interface {
	GetPrivacySettings(ctx context.Context, actor domain.Actor) (domain.PrivacySettings, error)
	UpdatePrivacySettings(ctx context.Context, input usecase.UpdatePrivacySettingsInput) (domain.PrivacySettings, error)
	GetRetentionSettings(ctx context.Context, actor domain.Actor) (domain.RetentionSettings, error)
	UpdateRetentionSettings(ctx context.Context, input usecase.UpdateRetentionSettingsInput) (domain.RetentionSettings, error)
	PreviewDeletion(ctx context.Context, input usecase.PreviewDeletionInput) (domain.DeletionPreview, error)
	CreateDeletionRequest(ctx context.Context, input usecase.CreateDeletionRequestInput) (domain.DeletionRequest, error)
	ListDeletionRequests(ctx context.Context, actor domain.Actor) ([]domain.DeletionRequest, error)
}

type Handler struct {
	privacy PrivacyUsecase
	logger  *slog.Logger
}

func NewHandler(privacy PrivacyUsecase, logger *slog.Logger) (*Handler, error) {
	if privacy == nil {
		return nil, errors.New("privacy usecase is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{privacy: privacy, logger: logger}, nil
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/analytics/privacy/settings", h.handlePrivacySettings)
	mux.HandleFunc("/api/v1/analytics/privacy/retention", h.handleRetentionSettings)
	mux.HandleFunc("/api/v1/analytics/privacy/deletion-preview", h.handleDeletionPreview)
	mux.HandleFunc("/api/v1/analytics/privacy/deletion-requests", h.handleDeletionRequests)
}

func (h *Handler) handlePrivacySettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		settings, err := h.privacy.GetPrivacySettings(r.Context(), actorFromRequest(r))
		if err != nil {
			h.writeUsecaseError(w, r, err)
			return
		}
		writeJSON(w, r, http.StatusOK, privacySettingsDTO(settings))
	case http.MethodPatch:
		var req updatePrivacySettingsRequest
		if err := decodeJSON(w, r, &req); err != nil {
			writeAPIError(w, r, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
			return
		}
		settings, err := h.privacy.UpdatePrivacySettings(r.Context(), usecase.UpdatePrivacySettingsInput{
			Actor:     actorFromRequest(r),
			Masking:   req.Masking,
			RequestID: requestID(r),
		})
		if err != nil {
			h.writeUsecaseError(w, r, err)
			return
		}
		writeJSON(w, r, http.StatusOK, privacySettingsDTO(settings))
	default:
		requireMethod(w, r, http.MethodGet, http.MethodPatch)
	}
}

func (h *Handler) handleRetentionSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		settings, err := h.privacy.GetRetentionSettings(r.Context(), actorFromRequest(r))
		if err != nil {
			h.writeUsecaseError(w, r, err)
			return
		}
		writeJSON(w, r, http.StatusOK, settings)
	case http.MethodPatch:
		var req updateRetentionSettingsRequest
		if err := decodeJSON(w, r, &req); err != nil {
			writeAPIError(w, r, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
			return
		}
		settings, err := h.privacy.UpdateRetentionSettings(r.Context(), usecase.UpdateRetentionSettingsInput{
			Actor:     actorFromRequest(r),
			Retention: retentionFromRequest(req),
			Reason:    req.Reason,
			RequestID: requestID(r),
		})
		if err != nil {
			h.writeUsecaseError(w, r, err)
			return
		}
		writeJSON(w, r, http.StatusOK, settings)
	default:
		requireMethod(w, r, http.MethodGet, http.MethodPatch)
	}
}

func (h *Handler) handleDeletionPreview(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var req deletionPreviewRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeAPIError(w, r, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		return
	}

	preview, err := h.privacy.PreviewDeletion(r.Context(), usecase.PreviewDeletionInput{
		Actor: actorFromRequest(r),
		Target: domain.DeletionTarget{
			Type:  req.TargetType,
			Value: req.TargetValue,
		},
	})
	if err != nil {
		h.writeUsecaseError(w, r, err)
		return
	}
	writeJSON(w, r, http.StatusOK, preview)
}

func (h *Handler) handleDeletionRequests(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		requests, err := h.privacy.ListDeletionRequests(r.Context(), actorFromRequest(r))
		if err != nil {
			h.writeUsecaseError(w, r, err)
			return
		}
		writeJSON(w, r, http.StatusOK, deletionRequestsResponse{Items: requests})
	case http.MethodPost:
		var req createDeletionRequest
		if err := decodeJSON(w, r, &req); err != nil {
			writeAPIError(w, r, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
			return
		}
		request, err := h.privacy.CreateDeletionRequest(r.Context(), usecase.CreateDeletionRequestInput{
			Actor: actorFromRequest(r),
			Target: domain.DeletionTarget{
				Type:  req.TargetType,
				Value: req.TargetValue,
			},
			Reason:    req.Reason,
			Confirmed: req.Confirmed,
			RequestID: requestID(r),
		})
		if err != nil {
			h.writeUsecaseError(w, r, err)
			return
		}
		writeJSON(w, r, http.StatusAccepted, request)
	default:
		requireMethod(w, r, http.MethodGet, http.MethodPost)
	}
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) error {
	defer r.Body.Close()
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain a single JSON object")
	}
	return nil
}

func requireMethod(w http.ResponseWriter, r *http.Request, methods ...string) bool {
	for _, method := range methods {
		if r.Method == method {
			return true
		}
	}
	w.Header().Set("Allow", strings.Join(methods, ", "))
	writeAPIError(w, r, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method is not allowed", nil)
	return false
}

func (h *Handler) writeUsecaseError(w http.ResponseWriter, r *http.Request, err error) {
	var fieldErr domain.FieldError
	switch {
	case errors.As(err, &fieldErr):
		writeAPIError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", fieldErr.Message, map[string]string{"field": fieldErr.Field})
	case errors.Is(err, domain.ErrUnauthenticated):
		writeAPIError(w, r, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "authentication is required", nil)
	case errors.Is(err, domain.ErrForbidden):
		writeAPIError(w, r, http.StatusForbidden, "PERMISSION_DENIED", "permission denied", nil)
	case errors.Is(err, domain.ErrReasonRequired):
		writeAPIError(w, r, http.StatusBadRequest, "REASON_REQUIRED", "audit reason must be at least 10 characters", nil)
	case errors.Is(err, domain.ErrConfirmation):
		writeAPIError(w, r, http.StatusBadRequest, "CONFIRMATION_REQUIRED", "deletion confirmation is required", nil)
	case errors.Is(err, domain.ErrNotFound):
		writeAPIError(w, r, http.StatusNotFound, "NOT_FOUND", "resource was not found", nil)
	default:
		h.logger.Error("session_privacy.http_error",
			slog.String("request_id", requestID(r)),
			slog.String("error", err.Error()),
		)
		writeAPIError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "privacy request failed", nil)
	}
}

func writeJSON(w http.ResponseWriter, r *http.Request, status int, data any) {
	rid := requestID(r)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", rid)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(apiEnvelope{
		Data:      data,
		RequestID: rid,
	})
}

func writeAPIError(w http.ResponseWriter, r *http.Request, status int, code string, message string, details any) {
	rid := requestID(r)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", rid)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(apiEnvelope{
		RequestID: rid,
		Error: &apiError{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

func actorFromRequest(r *http.Request) domain.Actor {
	actorID := firstHeader(r, "X-Admin-ID", "X-Actor-ID", "X-User-ID", "X-Authenticated-User-ID")
	return domain.Actor{
		ID:    strings.TrimSpace(actorID),
		Roles: splitRoles(firstHeader(r, "X-Admin-Roles", "X-Actor-Roles", "X-User-Roles", "X-Roles")),
	}
}

func requestID(r *http.Request) string {
	if value := firstHeader(r, "X-Request-ID", "X-Correlation-ID"); value != "" {
		return value
	}
	var bytes [12]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "req_unknown"
	}
	return "req_" + hex.EncodeToString(bytes[:])
}

func firstHeader(r *http.Request, names ...string) string {
	for _, name := range names {
		if value := strings.TrimSpace(r.Header.Get(name)); value != "" {
			return value
		}
	}
	return ""
}

func splitRoles(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	roles := make([]string, 0, len(parts))
	for _, part := range parts {
		role := strings.ToLower(strings.TrimSpace(part))
		if role != "" {
			roles = append(roles, role)
		}
	}
	return roles
}
