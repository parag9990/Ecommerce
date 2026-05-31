package domain

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

type PlatformSettingKey string

const (
	SettingMaintenanceMode PlatformSettingKey = "maintenance_mode"
	SettingCommissionRules PlatformSettingKey = "commission_rules"
	SettingFeatureFlags    PlatformSettingKey = "feature_flags"
	SettingSearchSynonyms  PlatformSettingKey = "search_synonyms"
)

func (k PlatformSettingKey) Valid() bool {
	_, ok := PlatformSettingDefinitionByKey(k)
	return ok
}

type SettingType string

const (
	SettingTypeMaintenance SettingType = "maintenance"
	SettingTypeCommission  SettingType = "commission"
	SettingTypeFeatureFlag SettingType = "feature_flags"
	SettingTypeSearch      SettingType = "search"
)

func (t SettingType) Valid() bool {
	switch t {
	case SettingTypeMaintenance, SettingTypeCommission, SettingTypeFeatureFlag, SettingTypeSearch:
		return true
	default:
		return false
	}
}

type PlatformSettingDefinition struct {
	Key  PlatformSettingKey
	Type SettingType
	Risk RiskLevel
}

var platformSettingDefinitions = []PlatformSettingDefinition{
	{Key: SettingMaintenanceMode, Type: SettingTypeMaintenance, Risk: RiskCritical},
	{Key: SettingCommissionRules, Type: SettingTypeCommission, Risk: RiskCritical},
	{Key: SettingFeatureFlags, Type: SettingTypeFeatureFlag, Risk: RiskHigh},
	{Key: SettingSearchSynonyms, Type: SettingTypeSearch, Risk: RiskMedium},
}

func PlatformSettingDefinitions() []PlatformSettingDefinition {
	out := make([]PlatformSettingDefinition, len(platformSettingDefinitions))
	copy(out, platformSettingDefinitions)
	return out
}

func KnownPlatformSettingKeys() []PlatformSettingKey {
	out := make([]PlatformSettingKey, 0, len(platformSettingDefinitions))
	for _, definition := range platformSettingDefinitions {
		out = append(out, definition.Key)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func PlatformSettingDefinitionByKey(key PlatformSettingKey) (PlatformSettingDefinition, bool) {
	for _, definition := range platformSettingDefinitions {
		if definition.Key == key {
			return definition, true
		}
	}
	return PlatformSettingDefinition{}, false
}

type PlatformSetting struct {
	Key              PlatformSettingKey `json:"key"`
	Type             SettingType        `json:"type"`
	Value            map[string]any     `json:"value"`
	Risk             RiskLevel          `json:"risk"`
	Version          uint64             `json:"version"`
	UpdatedByAdminID string             `json:"updated_by_admin_id,omitempty"`
	UpdateReason     string             `json:"update_reason,omitempty"`
	CreatedAt        time.Time          `json:"created_at,omitempty"`
	UpdatedAt        time.Time          `json:"updated_at,omitempty"`
}

func (s PlatformSetting) Clone() PlatformSetting {
	s.Value = CloneSettingValue(s.Value)
	return s
}

type PlatformSettingUpdateRequest struct {
	Value           map[string]any
	Reason          string
	ExpectedVersion uint64
}

type SearchSynonym struct {
	Root     string   `json:"root"`
	Synonyms []string `json:"synonyms"`
}

type SearchSynonymUpdateRequest struct {
	Root            string
	Synonyms        []string
	Reason          string
	ExpectedVersion uint64
}

type SearchSynonymsSnapshot struct {
	Synonyms  []SearchSynonym
	Version   uint64
	UpdatedAt time.Time
}

type PlatformSettingUpdatedEvent struct {
	EventID          string             `json:"event_id"`
	EventType        string             `json:"event_type"`
	SettingKey       PlatformSettingKey `json:"setting_key"`
	Version          uint64             `json:"version"`
	UpdatedByAdminID string             `json:"updated_by_admin_id"`
	UpdatedAt        time.Time          `json:"updated_at"`
}

func NewPlatformSettingUpdatedEvent(setting PlatformSetting) PlatformSettingUpdatedEvent {
	return PlatformSettingUpdatedEvent{
		EventID:          fmt.Sprintf("evt_setting_%s_v%d", setting.Key, setting.Version),
		EventType:        "PlatformSettingUpdated",
		SettingKey:       setting.Key,
		Version:          setting.Version,
		UpdatedByAdminID: setting.UpdatedByAdminID,
		UpdatedAt:        setting.UpdatedAt,
	}
}

func CloneSettingValue(value map[string]any) map[string]any {
	if value == nil {
		return nil
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		out := make(map[string]any, len(value))
		for key, item := range value {
			out[key] = item
		}
		return out
	}
	var out map[string]any
	if err := json.Unmarshal(encoded, &out); err != nil {
		out := make(map[string]any, len(value))
		for key, item := range value {
			out[key] = item
		}
		return out
	}
	return out
}
