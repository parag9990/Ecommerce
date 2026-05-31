package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"ecommerce/superadmin-service/internal/domain"
	"ecommerce/superadmin-service/internal/logging"
)

const DefaultPlatformSettingsCacheTTL = 5 * time.Minute

type PlatformSettingsRepository interface {
	ListSettings(ctx context.Context) ([]domain.PlatformSetting, error)
	GetSetting(ctx context.Context, key domain.PlatformSettingKey) (domain.PlatformSetting, error)
	UpdateSetting(ctx context.Context, setting domain.PlatformSetting, expectedVersion uint64) (domain.PlatformSetting, error)
}

type PlatformSettingsEventPublisher interface {
	PublishPlatformSettingUpdated(ctx context.Context, event domain.PlatformSettingUpdatedEvent) error
}

type PlatformSettingsConfig struct {
	CacheTTL time.Duration
}

type PlatformSettingsService struct {
	repo       PlatformSettingsRepository
	authorizer ControlAuthorizer
	audit      AuditRecorder
	publisher  PlatformSettingsEventPublisher
	cache      *platformSettingsCache
	logger     logging.Logger
}

func NewPlatformSettingsService(
	repo PlatformSettingsRepository,
	authorizer ControlAuthorizer,
	publisher PlatformSettingsEventPublisher,
	audit AuditRecorder,
	config PlatformSettingsConfig,
	logger logging.Logger,
) (*PlatformSettingsService, error) {
	if repo == nil {
		return nil, errors.New("platform settings service requires repository")
	}
	if authorizer == nil {
		return nil, errors.New("platform settings service requires authorizer")
	}
	if logger == nil {
		logger = logging.NewNop()
	}
	if audit == nil {
		audit = NewLoggingAuditRecorder(logger)
	}
	if config.CacheTTL == 0 {
		config.CacheTTL = DefaultPlatformSettingsCacheTTL
	}

	return &PlatformSettingsService{
		repo:       repo,
		authorizer: authorizer,
		audit:      audit,
		publisher:  publisher,
		cache:      newPlatformSettingsCache(config.CacheTTL),
		logger:     logger,
	}, nil
}

func (s *PlatformSettingsService) ListSettings(ctx context.Context) ([]domain.PlatformSetting, error) {
	actor, err := actorFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.authorizer.RequirePermission(ctx, actor, domain.PermissionSettingsRead); err != nil {
		return nil, err
	}

	if settings, ok := s.cache.GetList(); ok {
		return settings, nil
	}

	settings, err := s.repo.ListSettings(ctx)
	if err != nil {
		s.logger.Error(ctx, "platform settings list failed",
			"admin_id", actor.AdminID,
			"request_id", actor.RequestID,
			"error", err,
		)
		return nil, err
	}
	s.cache.SetList(settings)
	return clonePlatformSettings(settings), nil
}

func (s *PlatformSettingsService) UpdateSetting(ctx context.Context, key domain.PlatformSettingKey, req domain.PlatformSettingUpdateRequest) (domain.PlatformSetting, error) {
	actor, err := actorFromContext(ctx)
	if err != nil {
		return domain.PlatformSetting{}, err
	}
	reason, err := domain.NormalizeMutationReason(req.Reason)
	if err != nil {
		return domain.PlatformSetting{}, err
	}

	return s.updateSettingForActor(ctx, actor, key, req.Value, reason, req.ExpectedVersion)
}

func (s *PlatformSettingsService) ListSearchSynonyms(ctx context.Context) (domain.SearchSynonymsSnapshot, error) {
	actor, err := actorFromContext(ctx)
	if err != nil {
		return domain.SearchSynonymsSnapshot{}, err
	}
	if err := s.authorizer.RequirePermission(ctx, actor, domain.PermissionSearchSynonymsRead); err != nil {
		return domain.SearchSynonymsSnapshot{}, err
	}

	setting, err := s.getSetting(ctx, domain.SettingSearchSynonyms)
	if err != nil {
		return domain.SearchSynonymsSnapshot{}, err
	}
	synonyms, err := SearchSynonymsFromSettingValue(setting.Value)
	if err != nil {
		return domain.SearchSynonymsSnapshot{}, domain.NewInternal("stored search synonyms setting is invalid", err)
	}

	return domain.SearchSynonymsSnapshot{
		Synonyms:  synonyms,
		Version:   setting.Version,
		UpdatedAt: setting.UpdatedAt,
	}, nil
}

func (s *PlatformSettingsService) UpsertSearchSynonym(ctx context.Context, req domain.SearchSynonymUpdateRequest) (domain.SearchSynonymsSnapshot, error) {
	actor, err := actorFromContext(ctx)
	if err != nil {
		return domain.SearchSynonymsSnapshot{}, err
	}
	reason, err := domain.NormalizeMutationReason(req.Reason)
	if err != nil {
		return domain.SearchSynonymsSnapshot{}, err
	}
	if err := s.authorizer.RequireHighRiskPermission(ctx, actor, domain.PermissionSearchSynonymsWrite, reason); err != nil {
		return domain.SearchSynonymsSnapshot{}, err
	}

	current, err := s.getSetting(ctx, domain.SettingSearchSynonyms)
	if err != nil {
		return domain.SearchSynonymsSnapshot{}, err
	}
	if req.ExpectedVersion != 0 && req.ExpectedVersion != current.Version {
		return domain.SearchSynonymsSnapshot{}, domain.NewSettingVersionConflict(domain.SettingSearchSynonyms, req.ExpectedVersion, current.Version)
	}

	nextSynonym, err := NormalizeSearchSynonym(req.Root, req.Synonyms)
	if err != nil {
		return domain.SearchSynonymsSnapshot{}, err
	}
	synonyms, err := SearchSynonymsFromSettingValue(current.Value)
	if err != nil {
		return domain.SearchSynonymsSnapshot{}, domain.NewInternal("stored search synonyms setting is invalid", err)
	}

	replaced := false
	for i := range synonyms {
		if synonyms[i].Root == nextSynonym.Root {
			synonyms[i] = nextSynonym
			replaced = true
			break
		}
	}
	if !replaced {
		synonyms = append(synonyms, nextSynonym)
	}

	value := map[string]any{"synonyms": synonyms}
	updated, err := s.updateSettingForActorWithCurrent(ctx, actor, current, value, reason, req.ExpectedVersion)
	if err != nil {
		return domain.SearchSynonymsSnapshot{}, err
	}

	updatedSynonyms, err := SearchSynonymsFromSettingValue(updated.Value)
	if err != nil {
		return domain.SearchSynonymsSnapshot{}, domain.NewInternal("stored search synonyms setting is invalid after update", err)
	}
	return domain.SearchSynonymsSnapshot{
		Synonyms:  updatedSynonyms,
		Version:   updated.Version,
		UpdatedAt: updated.UpdatedAt,
	}, nil
}

func (s *PlatformSettingsService) updateSettingForActor(
	ctx context.Context,
	actor domain.AdminActor,
	key domain.PlatformSettingKey,
	value map[string]any,
	reason string,
	expectedVersion uint64,
) (domain.PlatformSetting, error) {
	if !key.Valid() {
		return domain.PlatformSetting{}, domain.NewValidationError(fmt.Sprintf("unsupported platform setting key %q", key))
	}
	permission := permissionForSettingUpdate(key)
	if err := s.authorizer.RequireHighRiskPermission(ctx, actor, permission, reason); err != nil {
		return domain.PlatformSetting{}, err
	}

	current, err := s.getSetting(ctx, key)
	if err != nil {
		return domain.PlatformSetting{}, err
	}
	return s.updateSettingForActorWithCurrent(ctx, actor, current, value, reason, expectedVersion)
}

func (s *PlatformSettingsService) updateSettingForActorWithCurrent(
	ctx context.Context,
	actor domain.AdminActor,
	current domain.PlatformSetting,
	value map[string]any,
	reason string,
	expectedVersion uint64,
) (domain.PlatformSetting, error) {
	if expectedVersion != 0 && expectedVersion != current.Version {
		return domain.PlatformSetting{}, domain.NewSettingVersionConflict(current.Key, expectedVersion, current.Version)
	}

	normalizedValue, err := NormalizePlatformSettingValue(current.Key, value)
	if err != nil {
		return domain.PlatformSetting{}, err
	}
	version := current.Version
	if expectedVersion != 0 {
		version = expectedVersion
	}

	next := current.Clone()
	next.Value = normalizedValue
	next.UpdatedByAdminID = actor.AdminID
	next.UpdateReason = reason

	updated, err := s.repo.UpdateSetting(ctx, next, version)
	if err != nil {
		s.logger.Error(ctx, "platform setting update failed",
			"admin_id", actor.AdminID,
			"setting_key", current.Key,
			"request_id", actor.RequestID,
			"error", err,
		)
		return domain.PlatformSetting{}, err
	}

	s.cache.InvalidateSetting(updated.Key)
	if err := s.recordSettingAudit(ctx, actor, current, updated, reason); err != nil {
		return domain.PlatformSetting{}, err
	}
	s.publishSettingUpdated(ctx, updated)
	return updated, nil
}

func (s *PlatformSettingsService) getSetting(ctx context.Context, key domain.PlatformSettingKey) (domain.PlatformSetting, error) {
	if setting, ok := s.cache.GetSetting(key); ok {
		return setting, nil
	}
	setting, err := s.repo.GetSetting(ctx, key)
	if err != nil {
		return domain.PlatformSetting{}, err
	}
	s.cache.SetSetting(setting)
	return setting.Clone(), nil
}

func (s *PlatformSettingsService) recordSettingAudit(ctx context.Context, actor domain.AdminActor, before domain.PlatformSetting, after domain.PlatformSetting, reason string) error {
	if err := s.audit.RecordAdminMutation(ctx, domain.AuditRecord{
		ActorAdminID: actor.AdminID,
		Action:       "platform_setting.update",
		ResourceType: "platform_setting",
		ResourceID:   string(after.Key),
		RequestID:    actor.RequestID,
		SessionID:    actor.SessionID,
		IPHash:       actor.IPHash,
		Reason:       reason,
		Before:       settingAuditSnapshot(before),
		After:        settingAuditSnapshot(after),
	}); err != nil {
		s.logger.Warn(ctx, "platform setting audit record failed",
			"setting_key", after.Key,
			"request_id", actor.RequestID,
			"error", err,
		)
		return auditRecordFailure(err)
	}
	return nil
}

func (s *PlatformSettingsService) publishSettingUpdated(ctx context.Context, setting domain.PlatformSetting) {
	if s.publisher == nil {
		return
	}
	event := domain.NewPlatformSettingUpdatedEvent(setting)
	if err := s.publisher.PublishPlatformSettingUpdated(ctx, event); err != nil {
		s.logger.Warn(ctx, "platform setting update event publish failed",
			"setting_key", setting.Key,
			"version", setting.Version,
			"event_id", event.EventID,
			"error", err,
		)
	}
}

func permissionForSettingUpdate(key domain.PlatformSettingKey) domain.Permission {
	if key == domain.SettingSearchSynonyms {
		return domain.PermissionSearchSynonymsWrite
	}
	return domain.PermissionSettingsWrite
}

func settingAuditSnapshot(setting domain.PlatformSetting) map[string]string {
	valueJSON, err := json.Marshal(setting.Value)
	if err != nil {
		valueJSON = []byte(`{}`)
	}
	return map[string]string{
		"key":     string(setting.Key),
		"type":    string(setting.Type),
		"risk":    string(setting.Risk),
		"version": strconv.FormatUint(setting.Version, 10),
		"value":   string(valueJSON),
	}
}

type cachedPlatformSetting struct {
	value     domain.PlatformSetting
	expiresAt time.Time
}

type cachedPlatformSettingList struct {
	values    []domain.PlatformSetting
	expiresAt time.Time
}

type platformSettingsCache struct {
	ttl    time.Duration
	mu     sync.RWMutex
	byKey  map[domain.PlatformSettingKey]cachedPlatformSetting
	list   cachedPlatformSettingList
	hasAll bool
}

func newPlatformSettingsCache(ttl time.Duration) *platformSettingsCache {
	if ttl < 0 {
		ttl = 0
	}
	return &platformSettingsCache{
		ttl:   ttl,
		byKey: make(map[domain.PlatformSettingKey]cachedPlatformSetting),
	}
}

func (c *platformSettingsCache) GetList() ([]domain.PlatformSetting, bool) {
	if c == nil || c.ttl == 0 {
		return nil, false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if !c.hasAll || time.Now().After(c.list.expiresAt) {
		return nil, false
	}
	return clonePlatformSettings(c.list.values), true
}

func (c *platformSettingsCache) SetList(settings []domain.PlatformSetting) {
	if c == nil || c.ttl == 0 {
		return
	}
	expiresAt := time.Now().Add(c.ttl)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.list = cachedPlatformSettingList{values: clonePlatformSettings(settings), expiresAt: expiresAt}
	c.hasAll = true
	for _, setting := range settings {
		c.byKey[setting.Key] = cachedPlatformSetting{value: setting.Clone(), expiresAt: expiresAt}
	}
}

func (c *platformSettingsCache) GetSetting(key domain.PlatformSettingKey) (domain.PlatformSetting, bool) {
	if c == nil || c.ttl == 0 {
		return domain.PlatformSetting{}, false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	cached, ok := c.byKey[key]
	if !ok || time.Now().After(cached.expiresAt) {
		return domain.PlatformSetting{}, false
	}
	return cached.value.Clone(), true
}

func (c *platformSettingsCache) SetSetting(setting domain.PlatformSetting) {
	if c == nil || c.ttl == 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.byKey[setting.Key] = cachedPlatformSetting{value: setting.Clone(), expiresAt: time.Now().Add(c.ttl)}
}

func (c *platformSettingsCache) InvalidateSetting(key domain.PlatformSettingKey) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.byKey, key)
	c.hasAll = false
}

func clonePlatformSettings(settings []domain.PlatformSetting) []domain.PlatformSetting {
	out := make([]domain.PlatformSetting, 0, len(settings))
	for _, setting := range settings {
		out = append(out, setting.Clone())
	}
	return out
}
