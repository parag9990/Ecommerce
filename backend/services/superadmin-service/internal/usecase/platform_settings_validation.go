package usecase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"ecommerce/superadmin-service/internal/domain"
)

const (
	maxMaintenanceMessageLength = 200
	maxCommissionRateBPS        = 5000
	maxFeatureDescription       = 300
	maxSynonymRootLength        = 80
	maxSynonymTermLength        = 80
	maxSynonymValuesPerRoot     = 20
	maxSearchSynonymEntries     = 500
)

var adminSettingIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_:-]{0,63}$`)

var knownFeatureFlags = map[string]struct{}{
	"coupon_engine":       {},
	"new_checkout":        {},
	"payment_retry":       {},
	"recommendations":     {},
	"search_autocomplete": {},
	"seller_bulk_upload":  {},
	"seller_kyc_v2":       {},
	"wishlist":            {},
}

var knownFeatureFlagRoles = map[string]struct{}{
	"admin":            {},
	"buyer":            {},
	"catalog_admin":    {},
	"finance_admin":    {},
	"operations_admin": {},
	"readonly_admin":   {},
	"seller":           {},
	"superadmin":       {},
}

func NormalizePlatformSettingValue(key domain.PlatformSettingKey, value map[string]any) (map[string]any, error) {
	if !key.Valid() {
		return nil, domain.NewValidationError(fmt.Sprintf("unsupported platform setting key %q", key))
	}
	if value == nil {
		return nil, domain.NewValidationError("setting value is required")
	}

	switch key {
	case domain.SettingMaintenanceMode:
		return normalizeMaintenanceMode(value)
	case domain.SettingCommissionRules:
		return normalizeCommissionRules(value)
	case domain.SettingFeatureFlags:
		return normalizeFeatureFlags(value)
	case domain.SettingSearchSynonyms:
		return normalizeSearchSynonyms(value)
	default:
		return nil, domain.NewValidationError(fmt.Sprintf("unsupported platform setting key %q", key))
	}
}

func NormalizeSearchSynonym(root string, synonyms []string) (domain.SearchSynonym, error) {
	cleanRoot := normalizeSettingTerm(root)
	if cleanRoot == "" {
		return domain.SearchSynonym{}, domain.NewValidationError("synonym root is required")
	}
	if len(cleanRoot) > maxSynonymRootLength {
		return domain.SearchSynonym{}, domain.NewValidationError(fmt.Sprintf("synonym root must be at most %d characters", maxSynonymRootLength))
	}
	if len(synonyms) == 0 {
		return domain.SearchSynonym{}, domain.NewValidationError("synonyms must contain at least one value")
	}
	if len(synonyms) > maxSynonymValuesPerRoot {
		return domain.SearchSynonym{}, domain.NewValidationError(fmt.Sprintf("synonyms can contain at most %d values", maxSynonymValuesPerRoot))
	}

	seen := make(map[string]struct{}, len(synonyms))
	cleanSynonyms := make([]string, 0, len(synonyms))
	for _, synonym := range synonyms {
		clean := normalizeSettingTerm(synonym)
		if clean == "" {
			return domain.SearchSynonym{}, domain.NewValidationError("synonym value cannot be empty")
		}
		if len(clean) > maxSynonymTermLength {
			return domain.SearchSynonym{}, domain.NewValidationError(fmt.Sprintf("synonym value must be at most %d characters", maxSynonymTermLength))
		}
		if clean == cleanRoot {
			return domain.SearchSynonym{}, domain.NewValidationError("synonym value cannot match root")
		}
		if _, ok := seen[clean]; ok {
			return domain.SearchSynonym{}, domain.NewValidationError("synonym values must be duplicate-free")
		}
		seen[clean] = struct{}{}
		cleanSynonyms = append(cleanSynonyms, clean)
	}
	sort.Strings(cleanSynonyms)

	return domain.SearchSynonym{Root: cleanRoot, Synonyms: cleanSynonyms}, nil
}

func SearchSynonymsFromSettingValue(value map[string]any) ([]domain.SearchSynonym, error) {
	normalized, err := normalizeSearchSynonyms(value)
	if err != nil {
		return nil, err
	}

	var payload searchSynonymsValue
	if err := decodeStrictSettingValue(normalized, &payload); err != nil {
		return nil, err
	}
	out := make([]domain.SearchSynonym, 0, len(payload.Synonyms))
	for _, item := range payload.Synonyms {
		out = append(out, domain.SearchSynonym{Root: item.Root, Synonyms: append([]string(nil), item.Synonyms...)})
	}
	return out, nil
}

type maintenanceModeValue struct {
	Enabled           *bool  `json:"enabled"`
	Message           string `json:"message"`
	StartsAt          string `json:"starts_at,omitempty"`
	EndsAt            string `json:"ends_at,omitempty"`
	AllowAdmins       *bool  `json:"allow_admins,omitempty"`
	AllowHealthChecks *bool  `json:"allow_health_checks,omitempty"`
}

func normalizeMaintenanceMode(value map[string]any) (map[string]any, error) {
	var payload maintenanceModeValue
	if err := decodeStrictSettingValue(value, &payload); err != nil {
		return nil, err
	}
	if payload.Enabled == nil {
		return nil, domain.NewValidationError("maintenance_mode.enabled is required")
	}

	payload.Message = strings.TrimSpace(payload.Message)
	if *payload.Enabled && payload.Message == "" {
		return nil, domain.NewValidationError("maintenance_mode.message is required when maintenance is enabled")
	}
	if len(payload.Message) > maxMaintenanceMessageLength {
		return nil, domain.NewValidationError(fmt.Sprintf("maintenance_mode.message must be at most %d characters", maxMaintenanceMessageLength))
	}
	if payload.AllowAdmins == nil {
		payload.AllowAdmins = boolPtr(true)
	}
	if payload.AllowHealthChecks == nil {
		payload.AllowHealthChecks = boolPtr(true)
	}
	if *payload.Enabled && (!*payload.AllowAdmins || !*payload.AllowHealthChecks) {
		return nil, domain.NewValidationError("maintenance mode must keep admin and health check access available")
	}

	startsAt, err := parseOptionalSettingTime(payload.StartsAt, "maintenance_mode.starts_at")
	if err != nil {
		return nil, err
	}
	endsAt, err := parseOptionalSettingTime(payload.EndsAt, "maintenance_mode.ends_at")
	if err != nil {
		return nil, err
	}
	if !startsAt.IsZero() && !endsAt.IsZero() && !endsAt.After(startsAt) {
		return nil, domain.NewValidationError("maintenance_mode.ends_at must be after starts_at")
	}

	payload.StartsAt = strings.TrimSpace(payload.StartsAt)
	payload.EndsAt = strings.TrimSpace(payload.EndsAt)
	return settingStructToMap(payload)
}

type commissionRulesValue struct {
	DefaultRateBPS   *int                         `json:"default_rate_bps"`
	Currency         string                       `json:"currency"`
	CategoryOverride []categoryCommissionOverride `json:"category_overrides"`
	SellerOverrides  []sellerCommissionOverride   `json:"seller_overrides"`
	EffectiveFrom    string                       `json:"effective_from"`
}

type categoryCommissionOverride struct {
	CategoryID string `json:"category_id"`
	RateBPS    *int   `json:"rate_bps"`
}

type sellerCommissionOverride struct {
	SellerID  string `json:"seller_id"`
	RateBPS   *int   `json:"rate_bps"`
	ExpiresAt string `json:"expires_at,omitempty"`
}

func normalizeCommissionRules(value map[string]any) (map[string]any, error) {
	var payload commissionRulesValue
	if err := decodeStrictSettingValue(value, &payload); err != nil {
		return nil, err
	}
	if payload.DefaultRateBPS == nil {
		return nil, domain.NewValidationError("commission_rules.default_rate_bps is required")
	}
	if err := validateCommissionRate(*payload.DefaultRateBPS, "commission_rules.default_rate_bps"); err != nil {
		return nil, err
	}

	payload.Currency = strings.ToUpper(strings.TrimSpace(payload.Currency))
	if payload.Currency != "INR" {
		return nil, domain.NewValidationError("commission_rules.currency must be INR")
	}
	if strings.TrimSpace(payload.EffectiveFrom) == "" {
		return nil, domain.NewValidationError("commission_rules.effective_from is required")
	}
	if _, err := parseRequiredSettingTime(payload.EffectiveFrom, "commission_rules.effective_from"); err != nil {
		return nil, err
	}
	payload.EffectiveFrom = strings.TrimSpace(payload.EffectiveFrom)

	if payload.CategoryOverride == nil {
		payload.CategoryOverride = []categoryCommissionOverride{}
	}
	if payload.SellerOverrides == nil {
		payload.SellerOverrides = []sellerCommissionOverride{}
	}

	seenCategories := make(map[string]struct{}, len(payload.CategoryOverride))
	for i := range payload.CategoryOverride {
		payload.CategoryOverride[i].CategoryID = strings.TrimSpace(payload.CategoryOverride[i].CategoryID)
		if !adminSettingIDPattern.MatchString(payload.CategoryOverride[i].CategoryID) {
			return nil, domain.NewValidationError("commission_rules.category_overrides.category_id has invalid format")
		}
		if payload.CategoryOverride[i].RateBPS == nil {
			return nil, domain.NewValidationError("commission_rules.category_overrides.rate_bps is required")
		}
		if err := validateCommissionRate(*payload.CategoryOverride[i].RateBPS, "commission_rules.category_overrides.rate_bps"); err != nil {
			return nil, err
		}
		if _, ok := seenCategories[payload.CategoryOverride[i].CategoryID]; ok {
			return nil, domain.NewValidationError("commission_rules.category_overrides must not contain duplicate category_id values")
		}
		seenCategories[payload.CategoryOverride[i].CategoryID] = struct{}{}
	}

	seenSellers := make(map[string]struct{}, len(payload.SellerOverrides))
	for i := range payload.SellerOverrides {
		payload.SellerOverrides[i].SellerID = strings.TrimSpace(payload.SellerOverrides[i].SellerID)
		if !adminSettingIDPattern.MatchString(payload.SellerOverrides[i].SellerID) {
			return nil, domain.NewValidationError("commission_rules.seller_overrides.seller_id has invalid format")
		}
		if payload.SellerOverrides[i].RateBPS == nil {
			return nil, domain.NewValidationError("commission_rules.seller_overrides.rate_bps is required")
		}
		if err := validateCommissionRate(*payload.SellerOverrides[i].RateBPS, "commission_rules.seller_overrides.rate_bps"); err != nil {
			return nil, err
		}
		payload.SellerOverrides[i].ExpiresAt = strings.TrimSpace(payload.SellerOverrides[i].ExpiresAt)
		if payload.SellerOverrides[i].ExpiresAt != "" {
			if _, err := parseRequiredSettingTime(payload.SellerOverrides[i].ExpiresAt, "commission_rules.seller_overrides.expires_at"); err != nil {
				return nil, err
			}
		}
		if _, ok := seenSellers[payload.SellerOverrides[i].SellerID]; ok {
			return nil, domain.NewValidationError("commission_rules.seller_overrides must not contain duplicate seller_id values")
		}
		seenSellers[payload.SellerOverrides[i].SellerID] = struct{}{}
	}

	return settingStructToMap(payload)
}

type featureFlagsValue struct {
	Flags map[string]featureFlagValue `json:"flags"`
}

type featureFlagValue struct {
	Enabled        *bool    `json:"enabled"`
	RolloutPercent *int     `json:"rollout_percent"`
	AllowedRoles   []string `json:"allowed_roles,omitempty"`
	Description    string   `json:"description,omitempty"`
}

func normalizeFeatureFlags(value map[string]any) (map[string]any, error) {
	var payload featureFlagsValue
	if err := decodeStrictSettingValue(value, &payload); err != nil {
		return nil, err
	}
	if payload.Flags == nil {
		return nil, domain.NewValidationError("feature_flags.flags is required")
	}
	if len(payload.Flags) == 0 {
		return nil, domain.NewValidationError("feature_flags.flags must contain at least one flag")
	}

	normalized := featureFlagsValue{Flags: make(map[string]featureFlagValue, len(payload.Flags))}
	for name, flag := range payload.Flags {
		cleanName := normalizeFlagName(name)
		if cleanName == "" {
			return nil, domain.NewValidationError("feature flag name is required")
		}
		if _, ok := knownFeatureFlags[cleanName]; !ok {
			return nil, domain.NewValidationError(fmt.Sprintf("unknown feature flag %q", cleanName))
		}
		if flag.Enabled == nil {
			return nil, domain.NewValidationError(fmt.Sprintf("feature flag %q enabled is required", cleanName))
		}
		if flag.RolloutPercent == nil {
			return nil, domain.NewValidationError(fmt.Sprintf("feature flag %q rollout_percent is required", cleanName))
		}
		if *flag.RolloutPercent < 0 || *flag.RolloutPercent > 100 {
			return nil, domain.NewValidationError("feature flag rollout_percent must be between 0 and 100")
		}
		flag.Description = strings.TrimSpace(flag.Description)
		if len(flag.Description) > maxFeatureDescription {
			return nil, domain.NewValidationError(fmt.Sprintf("feature flag description must be at most %d characters", maxFeatureDescription))
		}
		roles, err := normalizeFeatureFlagRoles(flag.AllowedRoles)
		if err != nil {
			return nil, err
		}
		flag.AllowedRoles = roles
		normalized.Flags[cleanName] = flag
	}

	return settingStructToMap(normalized)
}

type searchSynonymsValue struct {
	Synonyms []searchSynonymItemValue `json:"synonyms"`
}

type searchSynonymItemValue struct {
	Root     string   `json:"root"`
	Synonyms []string `json:"synonyms"`
}

func normalizeSearchSynonyms(value map[string]any) (map[string]any, error) {
	var payload searchSynonymsValue
	if err := decodeStrictSettingValue(value, &payload); err != nil {
		return nil, err
	}
	if payload.Synonyms == nil {
		payload.Synonyms = []searchSynonymItemValue{}
	}
	if len(payload.Synonyms) > maxSearchSynonymEntries {
		return nil, domain.NewValidationError(fmt.Sprintf("search_synonyms.synonyms can contain at most %d entries", maxSearchSynonymEntries))
	}

	seenRoots := make(map[string]struct{}, len(payload.Synonyms))
	normalized := searchSynonymsValue{Synonyms: make([]searchSynonymItemValue, 0, len(payload.Synonyms))}
	for _, item := range payload.Synonyms {
		synonym, err := NormalizeSearchSynonym(item.Root, item.Synonyms)
		if err != nil {
			return nil, err
		}
		if _, ok := seenRoots[synonym.Root]; ok {
			return nil, domain.NewValidationError("search_synonyms.synonyms must not contain duplicate roots")
		}
		seenRoots[synonym.Root] = struct{}{}
		normalized.Synonyms = append(normalized.Synonyms, searchSynonymItemValue{
			Root:     synonym.Root,
			Synonyms: synonym.Synonyms,
		})
	}
	sort.Slice(normalized.Synonyms, func(i, j int) bool {
		return normalized.Synonyms[i].Root < normalized.Synonyms[j].Root
	})

	return settingStructToMap(normalized)
}

func decodeStrictSettingValue(value map[string]any, target any) error {
	encoded, err := json.Marshal(value)
	if err != nil {
		return domain.NewValidationError("setting value must be JSON serializable")
	}
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return domain.NewValidationError("setting value does not match required schema")
	}
	if decoder.More() {
		return domain.NewValidationError("setting value must contain a single JSON object")
	}
	return nil
}

func settingStructToMap(value any) (map[string]any, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, domain.NewValidationError("setting value must be JSON serializable")
	}
	var out map[string]any
	if err := json.Unmarshal(encoded, &out); err != nil {
		return nil, domain.NewValidationError("setting value must be a JSON object")
	}
	return out, nil
}

func validateCommissionRate(rate int, field string) error {
	if rate < 0 || rate > maxCommissionRateBPS {
		return domain.NewValidationError(fmt.Sprintf("%s must be between 0 and %d bps", field, maxCommissionRateBPS))
	}
	return nil
}

func parseOptionalSettingTime(raw string, field string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, nil
	}
	return parseRequiredSettingTime(raw, field)
}

func parseRequiredSettingTime(raw string, field string) (time.Time, error) {
	value, err := time.Parse(time.RFC3339, strings.TrimSpace(raw))
	if err != nil {
		return time.Time{}, domain.NewValidationError(field + " must be an RFC3339 timestamp")
	}
	return value, nil
}

func normalizeSettingTerm(term string) string {
	term = strings.TrimSpace(strings.ToLower(term))
	return strings.Join(strings.Fields(term), " ")
}

func normalizeFlagName(name string) string {
	name = strings.TrimSpace(strings.ToLower(name))
	name = strings.ReplaceAll(name, "-", "_")
	return name
}

func normalizeFeatureFlagRoles(roles []string) ([]string, error) {
	if len(roles) == 0 {
		return []string{}, nil
	}

	seen := make(map[string]struct{}, len(roles))
	out := make([]string, 0, len(roles))
	for _, role := range roles {
		clean := strings.TrimSpace(strings.ToLower(role))
		if clean == "" {
			return nil, domain.NewValidationError("feature flag allowed_roles cannot contain empty values")
		}
		if _, ok := knownFeatureFlagRoles[clean]; !ok {
			return nil, domain.NewValidationError(fmt.Sprintf("unknown feature flag allowed role %q", clean))
		}
		if _, ok := seen[clean]; ok {
			continue
		}
		seen[clean] = struct{}{}
		out = append(out, clean)
	}
	sort.Strings(out)
	return out, nil
}

func boolPtr(value bool) *bool {
	return &value
}
