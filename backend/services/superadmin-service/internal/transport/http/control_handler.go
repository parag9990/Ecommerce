package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"ecommerce/superadmin-service/internal/domain"
	"ecommerce/superadmin-service/internal/logging"
)

type ControlUsecase interface {
	ListUsersForAdmin(ctx context.Context, req domain.AdminUserListRequest) (domain.AdminUserListResponse, error)
	UpdateUserStatus(ctx context.Context, userID string, req domain.StatusUpdateRequest) (domain.SuccessResponse, error)
	ListSellersForAdmin(ctx context.Context, req domain.AdminSellerListRequest) (domain.AdminSellerListResponse, error)
	UpdateSellerStatus(ctx context.Context, sellerID string, req domain.StatusUpdateRequest) (domain.SuccessResponse, error)
}

type ControlHandler struct {
	controls ControlUsecase
	logger   logging.Logger
}

func NewControlHandler(controls ControlUsecase, logger logging.Logger) *ControlHandler {
	if logger == nil {
		logger = logging.NewNop()
	}
	return &ControlHandler{controls: controls, logger: logger}
}

func (h *ControlHandler) Register(mux *http.ServeMux) {
	protected := ActorMiddleware
	mux.Handle("/api/v1/admin/users", protected(http.HandlerFunc(h.listUsers)))
	mux.Handle("/api/v1/admin/users/", protected(http.HandlerFunc(h.updateUserStatus)))
	mux.Handle("/api/v1/admin/sellers", protected(http.HandlerFunc(h.listSellers)))
	mux.Handle("/api/v1/admin/sellers/", protected(http.HandlerFunc(h.updateSellerStatus)))
}

func (h *ControlHandler) listUsers(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/admin/users" {
		writeError(w, r, domain.NewValidationError("expected /api/v1/admin/users"))
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Code: "METHOD_NOT_ALLOWED", Message: "method not allowed"})
		return
	}

	req, err := userListRequestFromQuery(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	response, err := h.controls.ListUsersForAdmin(r.Context(), req)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *ControlHandler) updateUserStatus(w http.ResponseWriter, r *http.Request) {
	userID, ok := resourceStatusID(r.URL.Path, "/api/v1/admin/users/")
	if !ok {
		writeError(w, r, domain.NewValidationError("expected /api/v1/admin/users/{user_id}/status"))
		return
	}
	if r.Method != http.MethodPatch {
		w.Header().Set("Allow", http.MethodPatch)
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Code: "METHOD_NOT_ALLOWED", Message: "method not allowed"})
		return
	}

	var req domain.StatusUpdateRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	response, err := h.controls.UpdateUserStatus(r.Context(), userID, req)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *ControlHandler) listSellers(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/admin/sellers" {
		writeError(w, r, domain.NewValidationError("expected /api/v1/admin/sellers"))
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Code: "METHOD_NOT_ALLOWED", Message: "method not allowed"})
		return
	}

	req, err := sellerListRequestFromQuery(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	response, err := h.controls.ListSellersForAdmin(r.Context(), req)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *ControlHandler) updateSellerStatus(w http.ResponseWriter, r *http.Request) {
	sellerID, ok := resourceStatusID(r.URL.Path, "/api/v1/admin/sellers/")
	if !ok {
		writeError(w, r, domain.NewValidationError("expected /api/v1/admin/sellers/{seller_id}/status"))
		return
	}
	if r.Method != http.MethodPatch {
		w.Header().Set("Allow", http.MethodPatch)
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Code: "METHOD_NOT_ALLOWED", Message: "method not allowed"})
		return
	}

	var req domain.StatusUpdateRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	response, err := h.controls.UpdateSellerStatus(r.Context(), sellerID, req)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func userListRequestFromQuery(r *http.Request) (domain.AdminUserListRequest, error) {
	query := r.URL.Query()
	page, err := positiveIntQuery(query.Get("page"), "page")
	if err != nil {
		return domain.AdminUserListRequest{}, err
	}
	pageSize, err := positiveIntQuery(query.Get("page_size"), "page_size")
	if err != nil {
		return domain.AdminUserListRequest{}, err
	}

	req := domain.AdminUserListRequest{
		Query:  query.Get("q"),
		Status: domain.UserStatus(strings.TrimSpace(query.Get("status"))),
		Pagination: domain.Pagination{
			Page:     page,
			PageSize: pageSize,
			Cursor:   query.Get("cursor"),
		},
	}
	return req.Normalize()
}

func sellerListRequestFromQuery(r *http.Request) (domain.AdminSellerListRequest, error) {
	query := r.URL.Query()
	page, err := positiveIntQuery(query.Get("page"), "page")
	if err != nil {
		return domain.AdminSellerListRequest{}, err
	}
	pageSize, err := positiveIntQuery(query.Get("page_size"), "page_size")
	if err != nil {
		return domain.AdminSellerListRequest{}, err
	}

	req := domain.AdminSellerListRequest{
		Query:  query.Get("q"),
		Status: domain.SellerStatus(strings.TrimSpace(query.Get("status"))),
		Pagination: domain.Pagination{
			Page:     page,
			PageSize: pageSize,
			Cursor:   query.Get("cursor"),
		},
	}
	return req.Normalize()
}

func positiveIntQuery(raw string, field string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return 0, domain.NewValidationError(field + " must be a positive integer")
	}
	return value, nil
}

func decodeJSONBody(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return domain.NewValidationError("invalid JSON request body")
	}
	var extra struct{}
	if err := decoder.Decode(&extra); err == nil {
		return domain.NewValidationError("request body must contain a single JSON object")
	} else if !errors.Is(err, io.EOF) {
		return domain.NewValidationError("invalid JSON request body")
	}
	return nil
}

func resourceStatusID(path string, prefix string) (string, bool) {
	if !strings.HasPrefix(path, prefix) {
		return "", false
	}
	remaining := strings.TrimPrefix(path, prefix)
	id, suffix, ok := strings.Cut(remaining, "/")
	if !ok || strings.TrimSpace(id) == "" || suffix != "status" {
		return "", false
	}
	unescaped, err := url.PathUnescape(id)
	if err != nil || strings.TrimSpace(unescaped) == "" {
		return "", false
	}
	return unescaped, true
}
