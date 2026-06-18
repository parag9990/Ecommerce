package usecase_test

import (
	"context"
	"testing"
	"time"

	"ecommerce/superadmin-service/internal/domain"
	"ecommerce/superadmin-service/internal/logging"
	"ecommerce/superadmin-service/internal/rbac"
	"ecommerce/superadmin-service/internal/usecase"
)

func TestPlatformSettingsUpdateMaintenanceMode(t *testing.T) {
	repo := newPlatformSettingsRepoStub()
	publisher := &recordingSettingsPublisher{}
	audit := &recordingAudit{}
	service := newPlatformSettingsService(t, repo, publisher, audit)

	updated, err := service.UpdateSetting(settingsActorContext("superadmin_1", domain.RoleSuperadmin, true), domain.SettingMaintenanceMode, domain.PlatformSettingUpdateRequest{
		Value: map[string]any{
			"enabled":             true,
			"message":             "Scheduled database maintenance",
			"starts_at":           "2026-06-01T01:00:00Z",
			"ends_at":             "2026-06-01T02:00:00Z",
			"allow_admins":        true,
			"allow_health_checks": true,
		},
		Reason:          "scheduled database maintenance",
		ExpectedVersion: 1,
	})
	if err != nil {
		t.Fatalf("UpdateSetting returned error: %v", err)
	}
	if updated.Version != 2 {
		t.Fatalf("version = %d, want 2", updated.Version)
	}
	if enabled, ok := updated.Value["enabled"].(bool); !ok || !enabled {
		t.Fatalf("enabled = %#v, want true", updated.Value["enabled"])
	}
	if len(audit.records) != 1 || audit.records[0].Action != "platform_setting.update" {
		t.Fatalf("audit records = %+v", audit.records)
	}
	if len(publisher.events) != 1 || publisher.events[0].SettingKey != domain.SettingMaintenanceMode {
		t.Fatalf("published events = %+v", publisher.events)
	}
}

func TestCatalogAdminCannotUpdateCommissionRules(t *testing.T) {
	service := newPlatformSettingsService(t, newPlatformSettingsRepoStub(), &recordingSettingsPublisher{}, &recordingAudit{})

	_, err := service.UpdateSetting(settingsActorContext("catalog_1", domain.RoleCatalog, false), domain.SettingCommissionRules, domain.PlatformSettingUpdateRequest{
		Value: map[string]any{
			"default_rate_bps": 1200,
			"currency":         "INR",
			"category_overrides": []map[string]any{
				{"category_id": "cat_electronics", "rate_bps": 900},
			},
			"seller_overrides": []map[string]any{},
			"effective_from":   "2026-06-15T00:00:00Z",
		},
		Reason:          "commission update requested by finance",
		ExpectedVersion: 1,
	})
	assertSettingsAppErrorCode(t, err, domain.CodeForbidden)
}

func TestCatalogAdminCanUpsertSearchSynonym(t *testing.T) {
	repo := newPlatformSettingsRepoStub()
	service := newPlatformSettingsService(t, repo, &recordingSettingsPublisher{}, &recordingAudit{})

	snapshot, err := service.UpsertSearchSynonym(settingsActorContext("catalog_1", domain.RoleCatalog, false), domain.SearchSynonymUpdateRequest{
		Root:            " Mobile ",
		Synonyms:        []string{"Phone", "smartphone"},
		Reason:          "catalog team synonym improvement",
		ExpectedVersion: 1,
	})
	if err != nil {
		t.Fatalf("UpsertSearchSynonym returned error: %v", err)
	}
	if snapshot.Version != 2 {
		t.Fatalf("version = %d, want 2", snapshot.Version)
	}
	if len(snapshot.Synonyms) != 1 {
		t.Fatalf("synonyms = %+v, want one entry", snapshot.Synonyms)
	}
	if snapshot.Synonyms[0].Root != "mobile" {
		t.Fatalf("root = %q, want mobile", snapshot.Synonyms[0].Root)
	}
	if got := snapshot.Synonyms[0].Synonyms; len(got) != 2 || got[0] != "phone" || got[1] != "smartphone" {
		t.Fatalf("synonyms = %+v", got)
	}
}

func TestPlatformSettingValidations(t *testing.T) {
	service := newPlatformSettingsService(t, newPlatformSettingsRepoStub(), &recordingSettingsPublisher{}, &recordingAudit{})
	ctx := settingsActorContext("superadmin_1", domain.RoleSuperadmin, true)

	_, err := service.UpdateSetting(ctx, domain.SettingCommissionRules, domain.PlatformSettingUpdateRequest{
		Value: map[string]any{
			"default_rate_bps":   6000,
			"currency":           "INR",
			"category_overrides": []map[string]any{},
			"seller_overrides":   []map[string]any{},
			"effective_from":     "2026-06-15T00:00:00Z",
		},
		Reason:          "commission update requested by finance",
		ExpectedVersion: 1,
	})
	assertSettingsAppErrorCode(t, err, domain.CodeValidationFailed)

	_, err = service.UpdateSetting(ctx, domain.SettingMaintenanceMode, domain.PlatformSettingUpdateRequest{
		Value: map[string]any{
			"enabled":             true,
			"message":             "Scheduled database maintenance",
			"starts_at":           "2026-06-01T02:00:00Z",
			"ends_at":             "2026-06-01T01:00:00Z",
			"allow_admins":        true,
			"allow_health_checks": true,
		},
		Reason:          "scheduled database maintenance",
		ExpectedVersion: 1,
	})
	assertSettingsAppErrorCode(t, err, domain.CodeValidationFailed)

	_, err = service.UpdateSetting(ctx, domain.SettingFeatureFlags, domain.PlatformSettingUpdateRequest{
		Value: map[string]any{
			"flags": map[string]any{
				"new_checkout": map[string]any{
					"enabled":         true,
					"rollout_percent": 101,
					"allowed_roles":   []string{"buyer"},
					"description":     "New checkout experience",
				},
			},
		},
		Reason:          "checkout beta rollout planning",
		ExpectedVersion: 1,
	})
	assertSettingsAppErrorCode(t, err, domain.CodeValidationFailed)

	_, err = usecase.NormalizeSearchSynonym("mobile", []string{"phone", " Phone "})
	assertSettingsAppErrorCode(t, err, domain.CodeValidationFailed)
}

func TestPlatformSettingVersionConflict(t *testing.T) {
	service := newPlatformSettingsService(t, newPlatformSettingsRepoStub(), &recordingSettingsPublisher{}, &recordingAudit{})

	_, err := service.UpdateSetting(settingsActorContext("superadmin_1", domain.RoleSuperadmin, true), domain.SettingMaintenanceMode, domain.PlatformSettingUpdateRequest{
		Value: map[string]any{
			"enabled":             false,
			"message":             "",
			"allow_admins":        true,
			"allow_health_checks": true,
		},
		Reason:          "cancel scheduled maintenance window",
		ExpectedVersion: 99,
	})
	assertSettingsAppErrorCode(t, err, domain.CodeSettingVersionConflict)
}

func newPlatformSettingsService(t *testing.T, repo usecase.PlatformSettingsRepository, publisher usecase.PlatformSettingsEventPublisher, audit usecase.AuditRecorder) *usecase.PlatformSettingsService {
	t.Helper()
	permissionRepo := rbac.NewStaticPermissionRepository([]rbac.StaticAdmin{
		{AdminID: "superadmin_1", Role: domain.RoleSuperadmin, Active: true},
		{AdminID: "catalog_1", Role: domain.RoleCatalog, Active: true},
		{AdminID: "readonly_1", Role: domain.RoleReadonly, Active: true},
	})
	authz, err := usecase.NewAuthorizationService(permissionRepo, logging.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	service, err := usecase.NewPlatformSettingsService(repo, authz, publisher, audit, usecase.PlatformSettingsConfig{CacheTTL: time.Minute}, logging.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	return service
}

func settingsActorContext(adminID string, role domain.AdminRole, mfa bool) context.Context {
	actor := domain.AdminActor{
		AdminID:     adminID,
		UserID:      "user_" + adminID,
		Roles:       []domain.AdminRole{role},
		SessionID:   "sess_1",
		RequestID:   "req_1",
		IPHash:      "hash_1",
		MFAVerified: mfa,
	}
	return domain.ContextWithActor(context.Background(), actor)
}

func assertSettingsAppErrorCode(t *testing.T, err error, want domain.ErrorCode) {
	t.Helper()
	appErr, ok := domain.AsAppError(err)
	if !ok || appErr.Code != want {
		t.Fatalf("error = %v, want %s", err, want)
	}
}

type platformSettingsRepoStub struct {
	settings map[domain.PlatformSettingKey]domain.PlatformSetting
}

func newPlatformSettingsRepoStub() *platformSettingsRepoStub {
	now := time.Date(2026, 5, 31, 10, 0, 0, 0, time.UTC)
	settings := make(map[domain.PlatformSettingKey]domain.PlatformSetting)
	for _, definition := range domain.PlatformSettingDefinitions() {
		settings[definition.Key] = domain.PlatformSetting{
			Key:              definition.Key,
			Type:             definition.Type,
			Risk:             definition.Risk,
			Value:            defaultSettingValue(definition.Key),
			Version:          1,
			UpdatedByAdminID: "system",
			UpdateReason:     "initial default",
			CreatedAt:        now,
			UpdatedAt:        now,
		}
	}
	return &platformSettingsRepoStub{settings: settings}
}

func (r *platformSettingsRepoStub) ListSettings(ctx context.Context) ([]domain.PlatformSetting, error) {
	out := make([]domain.PlatformSetting, 0, len(r.settings))
	for _, key := range domain.KnownPlatformSettingKeys() {
		out = append(out, r.settings[key].Clone())
	}
	return out, nil
}

func (r *platformSettingsRepoStub) GetSetting(ctx context.Context, key domain.PlatformSettingKey) (domain.PlatformSetting, error) {
	setting, ok := r.settings[key]
	if !ok {
		return domain.PlatformSetting{}, domain.NewPlatformSettingNotFound(key)
	}
	return setting.Clone(), nil
}

func (r *platformSettingsRepoStub) UpdateSetting(ctx context.Context, setting domain.PlatformSetting, expectedVersion uint64) (domain.PlatformSetting, error) {
	current, ok := r.settings[setting.Key]
	if !ok {
		return domain.PlatformSetting{}, domain.NewPlatformSettingNotFound(setting.Key)
	}
	if current.Version != expectedVersion {
		return domain.PlatformSetting{}, domain.NewSettingVersionConflict(setting.Key, expectedVersion, current.Version)
	}
	setting.Version = current.Version + 1
	setting.CreatedAt = current.CreatedAt
	setting.UpdatedAt = current.UpdatedAt.Add(time.Minute)
	r.settings[setting.Key] = setting.Clone()
	return setting.Clone(), nil
}

func defaultSettingValue(key domain.PlatformSettingKey) map[string]any {
	switch key {
	case domain.SettingMaintenanceMode:
		return map[string]any{
			"enabled":             false,
			"message":             "",
			"allow_admins":        true,
			"allow_health_checks": true,
		}
	case domain.SettingCommissionRules:
		return map[string]any{
			"default_rate_bps":   1000,
			"currency":           "INR",
			"category_overrides": []any{},
			"seller_overrides":   []any{},
			"effective_from":     "2026-01-01T00:00:00Z",
		}
	case domain.SettingFeatureFlags:
		return map[string]any{
			"flags": map[string]any{
				"new_checkout": map[string]any{
					"enabled":         false,
					"rollout_percent": 0,
					"allowed_roles":   []string{"buyer"},
					"description":     "New checkout experience",
				},
			},
		}
	case domain.SettingSearchSynonyms:
		return map[string]any{"synonyms": []any{}}
	default:
		return map[string]any{}
	}
}

type recordingSettingsPublisher struct {
	events []domain.PlatformSettingUpdatedEvent
}

func (p *recordingSettingsPublisher) PublishPlatformSettingUpdated(ctx context.Context, event domain.PlatformSettingUpdatedEvent) error {
	p.events = append(p.events, event)
	return nil
}
