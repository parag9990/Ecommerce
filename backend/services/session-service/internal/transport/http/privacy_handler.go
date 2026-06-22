package httptransport

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/example/ecommerce-platform/backend/services/session-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/session-service/internal/usecase"
)

type updatePrivacySettingsRequest struct {
	Masking domain.PrivacyMaskingSettings `json:"masking"`
}

type updateRetentionSettingsRequest struct {
	domain.RetentionSettings
	Reason string `json:"reason"`
}

type deletionTargetRequest struct {
	TargetType  domain.DeletionTargetType `json:"targetType"`
	TargetValue string                    `json:"targetValue"`
}

type createDeletionRequest struct {
	deletionTargetRequest
	Reason    string `json:"reason"`
	Confirmed bool   `json:"confirmed"`
}

type deletionRequestsResponse struct {
	Items []domain.DeletionRequest `json:"items"`
}

func (h *Handler) handlePrivacySettings(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdminAccess(w, r) {
		return
	}
	if h.privacy == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "PRIVACY_UNAVAILABLE", "Privacy service is not configured")
		return
	}

	actor := actorFromRequest(r)
	switch r.Method {
	case http.MethodGet:
		out, err := h.privacy.GetPrivacySettings(r.Context(), actor)
		if err != nil {
			h.writePrivacyError(w, r, err)
			return
		}
		writeJSON(w, http.StatusOK, out)
	case http.MethodPatch:
		if !requireJSONRequest(w, r) {
			return
		}
		var req updatePrivacySettingsRequest
		if err := h.decodeJSON(w, r, &req); err != nil {
			writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
			return
		}
		out, err := h.privacy.UpdatePrivacySettings(r.Context(), usecase.UpdatePrivacySettingsInput{
			Actor:     actor,
			Masking:   req.Masking,
			RequestID: requestID(r),
		})
		if err != nil {
			h.writePrivacyError(w, r, err)
			return
		}
		writeJSON(w, http.StatusOK, out)
	default:
		writeMethodNotAllowed(w, http.MethodGet+", "+http.MethodPatch)
	}
}

func (h *Handler) handlePrivacyRetention(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdminAccess(w, r) {
		return
	}
	if h.privacy == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "PRIVACY_UNAVAILABLE", "Privacy service is not configured")
		return
	}

	actor := actorFromRequest(r)
	switch r.Method {
	case http.MethodGet:
		out, err := h.privacy.GetRetentionSettings(r.Context(), actor)
		if err != nil {
			h.writePrivacyError(w, r, err)
			return
		}
		writeJSON(w, http.StatusOK, out)
	case http.MethodPatch:
		if !requireJSONRequest(w, r) {
			return
		}
		var req updateRetentionSettingsRequest
		if err := h.decodeJSON(w, r, &req); err != nil {
			writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
			return
		}
		out, err := h.privacy.UpdateRetentionSettings(r.Context(), usecase.UpdateRetentionSettingsInput{
			Actor:     actor,
			Retention: req.RetentionSettings,
			Reason:    req.Reason,
			RequestID: requestID(r),
		})
		if err != nil {
			h.writePrivacyError(w, r, err)
			return
		}
		writeJSON(w, http.StatusOK, out)
	default:
		writeMethodNotAllowed(w, http.MethodGet+", "+http.MethodPatch)
	}
}

func (h *Handler) handleDeletionPreview(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) || !h.requireAdminAccess(w, r) {
		return
	}
	if h.privacy == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "PRIVACY_UNAVAILABLE", "Privacy service is not configured")
		return
	}
	if !requireJSONRequest(w, r) {
		return
	}
	var req deletionTargetRequest
	if err := h.decodeJSON(w, r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	out, err := h.privacy.PreviewDeletion(r.Context(), usecase.PreviewDeletionInput{
		Actor: actorFromRequest(r),
		Target: domain.DeletionTarget{
			Type:  req.TargetType,
			Value: req.TargetValue,
		},
	})
	if err != nil {
		h.writePrivacyError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) handleDeletionRequests(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdminAccess(w, r) {
		return
	}
	if h.privacy == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "PRIVACY_UNAVAILABLE", "Privacy service is not configured")
		return
	}
	actor := actorFromRequest(r)
	switch r.Method {
	case http.MethodGet:
		items, err := h.privacy.ListDeletionRequests(r.Context(), actor)
		if err != nil {
			h.writePrivacyError(w, r, err)
			return
		}
		if items == nil {
			items = []domain.DeletionRequest{}
		}
		writeJSON(w, http.StatusOK, deletionRequestsResponse{Items: items})
	case http.MethodPost:
		if !requireJSONRequest(w, r) {
			return
		}
		var req createDeletionRequest
		if err := h.decodeJSON(w, r, &req); err != nil {
			writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
			return
		}
		out, err := h.privacy.CreateDeletionRequest(r.Context(), usecase.CreateDeletionRequestInput{
			Actor: actor,
			Target: domain.DeletionTarget{
				Type:  req.TargetType,
				Value: req.TargetValue,
			},
			Reason:    req.Reason,
			Confirmed: req.Confirmed,
			RequestID: requestID(r),
		})
		if err != nil {
			h.writePrivacyError(w, r, err)
			return
		}
		writeJSON(w, http.StatusCreated, out)
	default:
		writeMethodNotAllowed(w, http.MethodGet+", "+http.MethodPost)
	}
}

func actorFromRequest(r *http.Request) domain.Actor {
	return domain.Actor{ID: authenticatedActorID(r), Roles: rolesFromRequest(r)}
}

func requireJSONRequest(w http.ResponseWriter, r *http.Request) bool {
	if isJSONContentType(r.Header.Get("Content-Type")) {
		return true
	}
	writeAPIError(w, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", "Content-Type must be application/json")
	return false
}

func writeMethodNotAllowed(w http.ResponseWriter, allow string) {
	w.Header().Set("Allow", allow)
	writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
}

func (h *Handler) writePrivacyError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrUnauthenticated):
		writeAPIError(w, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "Authentication required")
	case errors.Is(err, domain.ErrForbidden):
		writeAPIError(w, http.StatusForbidden, "PERMISSION_DENIED", "Permission denied")
	case errors.Is(err, domain.ErrInvalidInput), errors.Is(err, domain.ErrReasonRequired), errors.Is(err, domain.ErrConfirmation):
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	case errors.Is(err, domain.ErrNotFound):
		writeAPIError(w, http.StatusNotFound, "NOT_FOUND", "Resource not found")
	default:
		h.logger.ErrorContext(r.Context(), "session.http.privacy_failed",
			slog.String("request_id", requestID(r)),
			slog.String("error", err.Error()),
		)
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
	}
}
