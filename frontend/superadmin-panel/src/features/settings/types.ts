export const PLATFORM_SETTING_KEYS = [
  "commission.default_rate",
  "commission.category_overrides",
  "platform.feature_flags",
  "platform.maintenance_mode"
] as const;

export type SettingKey = (typeof PLATFORM_SETTING_KEYS)[number];

export type PlatformSetting<TValue = unknown> = {
  key: SettingKey;
  value: TValue;
  updated_at: string | null;
  version?: number | null;
};

export type PlatformSettingsResponse = {
  settings: PlatformSetting[];
};

export type PlatformSettingInput<TValue> = {
  value: TValue;
  reason: string;
};

export type CommissionDefaultRate = {
  rate_percent: number;
  applies_to: "all_sellers";
};

export type CommissionCategoryOverride = {
  category_id: string;
  rate_percent: number;
};

export type CommissionCategoryOverridesValue = {
  overrides: CommissionCategoryOverride[];
};

export type FeatureFlagsValue = {
  flags: Record<string, boolean>;
};

export type MaintenanceModeValue = {
  enabled: boolean;
  message: string;
  starts_at: string | null;
  ends_at: string | null;
  allow_admin_bypass: boolean;
};

export type MaintenanceModeStatus = "disabled" | "scheduled" | "active" | "expired";

export type SearchSynonym = {
  synonym_id: string;
  root: string;
  synonyms: string[];
  updated_at?: string | null;
};

export type SearchSynonymInput = {
  root: string;
  synonyms: string[];
};

export type SearchSynonymMutationInput = SearchSynonymInput & {
  reason: string;
};

export type SearchSynonymListResponse = {
  synonyms: SearchSynonym[];
};
