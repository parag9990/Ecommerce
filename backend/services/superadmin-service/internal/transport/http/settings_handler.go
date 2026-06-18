package http

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	"ecommerce/superadmin-service/internal/domain"
	"ecommerce/superadmin-service/internal/logging"
)

type PlatformSettingsUsecase interface {
	ListSettings(ctx context.Context) ([]domain.PlatformSetting, error)
	UpdateSetting(ctx context.Context, key domain.PlatformSettingKey, req domain.PlatformSettingUpdateRequest) (domain.PlatformSetting, error)
	ListSearchSynonyms(ctx context.Context) (domain.SearchSynonymsSnapshot, error)
	UpsertSearchSynonym(ctx context.Context, req domain.SearchSynonymUpdateRequest) (domain.SearchSynonymsSnapshot, error)
}

type SettingsHandler struct {
	settings PlatformSettingsUsecase
	logger   logging.Logger
}

func NewSettingsHandler(settings PlatformSettingsUsecase, logger logging.Logger) *SettingsHandler {
	if logger == nil {
		logger = logging.NewNop()
	}
	return &SettingsHandler{settings: settings, logger: logger}
}

func (h *SettingsHandler) Register(mux *http.ServeMux) {
	protected := ActorMiddleware
	mux.Handle("/api/v1/admin/settings", protected(http.HandlerFunc(h.listSettings)))
	mux.Handle("/api/v1/admin/settings/", protected(http.HandlerFunc(h.updateSetting)))
	mux.Handle("/api/v1/admin/search/synonyms", protected(http.HandlerFunc(h.searchSynonyms)))
}

func (h *SettingsHandler) listSettings(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/admin/settings" {
		writeError(w, r, domain.NewValidationError("expected /api/v1/admin/settings"))
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Code: "METHOD_NOT_ALLOWED", Message: "method not allowed"})
		return
	}

	settings, err := h.settings.ListSettings(r.Context())
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, platformSettingsResponse{Settings: platformSettingDTOs(settings)})
}

func (h *SettingsHandler) updateSetting(w http.ResponseWriter, r *http.Request) {
	key, ok := settingKeyFromPath(r.URL.Path)
	if !ok {
		writeError(w, r, domain.NewValidationError("expected /api/v1/admin/settings/{key}"))
		return
	}
	if r.Method != http.MethodPatch {
		w.Header().Set("Allow", http.MethodPatch)
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Code: "METHOD_NOT_ALLOWED", Message: "method not allowed"})
		return
	}

	var input platformSettingInput
	if err := decodeJSONBody(r, &input); err != nil {
		writeError(w, r, err)
		return
	}

	setting, err := h.settings.UpdateSetting(r.Context(), key, domain.PlatformSettingUpdateRequest{
		Value:           input.Value,
		Reason:          input.Reason,
		ExpectedVersion: input.Version,
	})
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, platformSettingDTOFromDomain(setting))
}

func (h *SettingsHandler) searchSynonyms(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/admin/search/synonyms" {
		writeError(w, r, domain.NewValidationError("expected /api/v1/admin/search/synonyms"))
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.listSearchSynonyms(w, r)
	case http.MethodPost:
		h.upsertSearchSynonym(w, r)
	default:
		w.Header().Set("Allow", strings.Join([]string{http.MethodGet, http.MethodPost}, ", "))
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Code: "METHOD_NOT_ALLOWED", Message: "method not allowed"})
	}
}

func (h *SettingsHandler) listSearchSynonyms(w http.ResponseWriter, r *http.Request) {
	snapshot, err := h.settings.ListSearchSynonyms(r.Context())
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, searchSynonymListResponseFromDomain(snapshot))
}

func (h *SettingsHandler) upsertSearchSynonym(w http.ResponseWriter, r *http.Request) {
	var input searchSynonymInput
	if err := decodeJSONBody(r, &input); err != nil {
		writeError(w, r, err)
		return
	}

	snapshot, err := h.settings.UpsertSearchSynonym(r.Context(), domain.SearchSynonymUpdateRequest{
		Root:            input.Root,
		Synonyms:        input.Synonyms,
		Reason:          input.Reason,
		ExpectedVersion: input.Version,
	})
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, searchSynonymListResponseFromDomain(snapshot))
}

func settingKeyFromPath(path string) (domain.PlatformSettingKey, bool) {
	const prefix = "/api/v1/admin/settings/"
	if !strings.HasPrefix(path, prefix) {
		return "", false
	}
	raw := strings.TrimPrefix(path, prefix)
	if raw == "" || strings.Contains(raw, "/") {
		return "", false
	}
	unescaped, err := url.PathUnescape(raw)
	if err != nil || strings.TrimSpace(unescaped) == "" {
		return "", false
	}
	return domain.PlatformSettingKey(strings.TrimSpace(unescaped)), true
}

type platformSettingInput struct {
	Value   map[string]any `json:"value"`
	Reason  string         `json:"reason"`
	Version uint64         `json:"version,omitempty"`
}

type platformSettingDTO struct {
	Key       string         `json:"key"`
	Type      string         `json:"type,omitempty"`
	Risk      string         `json:"risk,omitempty"`
	Value     map[string]any `json:"value"`
	Version   uint64         `json:"version"`
	UpdatedAt string         `json:"updated_at,omitempty"`
}

type platformSettingsResponse struct {
	Settings []platformSettingDTO `json:"settings"`
}

type searchSynonymInput struct {
	Root     string   `json:"root"`
	Synonyms []string `json:"synonyms"`
	Reason   string   `json:"reason"`
	Version  uint64   `json:"version,omitempty"`
}

type searchSynonymDTO struct {
	SynonymID string   `json:"synonym_id"`
	Root      string   `json:"root"`
	Synonyms  []string `json:"synonyms"`
}

type searchSynonymListResponse struct {
	Synonyms  []searchSynonymDTO `json:"synonyms"`
	Version   uint64             `json:"version"`
	UpdatedAt string             `json:"updated_at,omitempty"`
}

func platformSettingDTOs(settings []domain.PlatformSetting) []platformSettingDTO {
	out := make([]platformSettingDTO, 0, len(settings))
	for _, setting := range settings {
		out = append(out, platformSettingDTOFromDomain(setting))
	}
	return out
}

func platformSettingDTOFromDomain(setting domain.PlatformSetting) platformSettingDTO {
	return platformSettingDTO{
		Key:       string(setting.Key),
		Type:      string(setting.Type),
		Risk:      string(setting.Risk),
		Value:     domain.CloneSettingValue(setting.Value),
		Version:   setting.Version,
		UpdatedAt: formatOptionalTime(setting.UpdatedAt),
	}
}

func searchSynonymListResponseFromDomain(snapshot domain.SearchSynonymsSnapshot) searchSynonymListResponse {
	out := searchSynonymListResponse{
		Synonyms:  make([]searchSynonymDTO, 0, len(snapshot.Synonyms)),
		Version:   snapshot.Version,
		UpdatedAt: formatOptionalTime(snapshot.UpdatedAt),
	}
	for _, synonym := range snapshot.Synonyms {
		out.Synonyms = append(out.Synonyms, searchSynonymDTO{
			SynonymID: searchSynonymID(synonym.Root),
			Root:      synonym.Root,
			Synonyms:  append([]string(nil), synonym.Synonyms...),
		})
	}
	return out
}

func searchSynonymID(root string) string {
	id := strings.TrimSpace(strings.ToLower(root))
	id = strings.ReplaceAll(id, " ", "_")
	return id
}

func formatOptionalTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}
