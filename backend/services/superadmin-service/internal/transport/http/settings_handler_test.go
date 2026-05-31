package http_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ecommerce/superadmin-service/internal/domain"
	"ecommerce/superadmin-service/internal/logging"
	transporthttp "ecommerce/superadmin-service/internal/transport/http"
)

func TestListSettingsHandler(t *testing.T) {
	settings := &settingsUsecaseStub{
		settings: []domain.PlatformSetting{
			{
				Key:       domain.SettingMaintenanceMode,
				Type:      domain.SettingTypeMaintenance,
				Risk:      domain.RiskCritical,
				Value:     map[string]any{"enabled": false, "message": "", "allow_admins": true, "allow_health_checks": true},
				Version:   1,
				UpdatedAt: time.Date(2026, 5, 31, 10, 0, 0, 0, time.UTC),
			},
		},
	}
	mux := settingsMux(settings)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/settings", nil)
	addAdminHeaders(req, "superadmin_1", domain.RoleSuperadmin)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var body map[string][]map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body["settings"]) != 1 || body["settings"][0]["key"] != string(domain.SettingMaintenanceMode) {
		t.Fatalf("body = %+v", body)
	}
}

func TestUpdateSettingHandler(t *testing.T) {
	settings := &settingsUsecaseStub{}
	mux := settingsMux(settings)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/settings/maintenance_mode", strings.NewReader(`{
		"value":{"enabled":false,"message":"","allow_admins":true,"allow_health_checks":true},
		"reason":"cancel scheduled maintenance",
		"version":3
	}`))
	addAdminHeaders(req, "superadmin_1", domain.RoleSuperadmin)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if settings.updatedKey != domain.SettingMaintenanceMode {
		t.Fatalf("updated key = %s", settings.updatedKey)
	}
	if settings.updateReq.ExpectedVersion != 3 || settings.updateReq.Reason != "cancel scheduled maintenance" {
		t.Fatalf("update request = %+v", settings.updateReq)
	}
}

func TestSearchSynonymsHandler(t *testing.T) {
	settings := &settingsUsecaseStub{}
	mux := settingsMux(settings)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/search/synonyms", strings.NewReader(`{
		"root":"mobile",
		"synonyms":["phone","smartphone"],
		"reason":"catalog synonym improvement",
		"version":1
	}`))
	addAdminHeaders(req, "catalog_1", domain.RoleCatalog)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if settings.synonymReq.Root != "mobile" || settings.synonymReq.ExpectedVersion != 1 {
		t.Fatalf("synonym request = %+v", settings.synonymReq)
	}
}

func settingsMux(settings *settingsUsecaseStub) *http.ServeMux {
	rbacHandler := transporthttp.NewRBACHandler(&authorizationStub{}, nil, logging.NewNop())
	settingsHandler := transporthttp.NewSettingsHandler(settings, logging.NewNop())
	return transporthttp.NewServeMux(rbacHandler, settingsHandler)
}

type settingsUsecaseStub struct {
	settings   []domain.PlatformSetting
	updatedKey domain.PlatformSettingKey
	updateReq  domain.PlatformSettingUpdateRequest
	synonymReq domain.SearchSynonymUpdateRequest
}

func (s *settingsUsecaseStub) ListSettings(ctx context.Context) ([]domain.PlatformSetting, error) {
	return s.settings, nil
}

func (s *settingsUsecaseStub) UpdateSetting(ctx context.Context, key domain.PlatformSettingKey, req domain.PlatformSettingUpdateRequest) (domain.PlatformSetting, error) {
	s.updatedKey = key
	s.updateReq = req
	definition, _ := domain.PlatformSettingDefinitionByKey(key)
	return domain.PlatformSetting{
		Key:       key,
		Type:      definition.Type,
		Risk:      definition.Risk,
		Value:     req.Value,
		Version:   req.ExpectedVersion + 1,
		UpdatedAt: time.Date(2026, 5, 31, 11, 0, 0, 0, time.UTC),
	}, nil
}

func (s *settingsUsecaseStub) ListSearchSynonyms(ctx context.Context) (domain.SearchSynonymsSnapshot, error) {
	return domain.SearchSynonymsSnapshot{
		Synonyms:  []domain.SearchSynonym{{Root: "mobile", Synonyms: []string{"phone"}}},
		Version:   1,
		UpdatedAt: time.Date(2026, 5, 31, 10, 0, 0, 0, time.UTC),
	}, nil
}

func (s *settingsUsecaseStub) UpsertSearchSynonym(ctx context.Context, req domain.SearchSynonymUpdateRequest) (domain.SearchSynonymsSnapshot, error) {
	s.synonymReq = req
	return domain.SearchSynonymsSnapshot{
		Synonyms:  []domain.SearchSynonym{{Root: req.Root, Synonyms: req.Synonyms}},
		Version:   req.ExpectedVersion + 1,
		UpdatedAt: time.Date(2026, 5, 31, 11, 0, 0, 0, time.UTC),
	}, nil
}
