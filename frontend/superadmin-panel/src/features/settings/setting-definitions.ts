import type {
  CommissionCategoryOverridesValue,
  CommissionDefaultRate,
  FeatureFlagsValue,
  MaintenanceModeValue,
  SettingKey
} from "./types";

export type SettingDefinition<TValue = unknown> = {
  key: SettingKey;
  label: string;
  description: string;
  defaultValue: TValue;
};

export const DEFAULT_COMMISSION_RATE: CommissionDefaultRate = {
  rate_percent: 12,
  applies_to: "all_sellers"
};

export const DEFAULT_COMMISSION_OVERRIDES: CommissionCategoryOverridesValue = {
  overrides: []
};

export const DEFAULT_FEATURE_FLAGS: FeatureFlagsValue = {
  flags: {
    new_checkout: false,
    seller_live_chat: false
  }
};

export const DEFAULT_MAINTENANCE_MODE: MaintenanceModeValue = {
  enabled: false,
  message: "",
  starts_at: null,
  ends_at: null,
  allow_admin_bypass: true
};

export const settingDefinitions: Record<SettingKey, SettingDefinition> = {
  "commission.default_rate": {
    key: "commission.default_rate",
    label: "Default commission",
    description: "Platform-wide seller commission percentage.",
    defaultValue: DEFAULT_COMMISSION_RATE
  },
  "commission.category_overrides": {
    key: "commission.category_overrides",
    label: "Category commission overrides",
    description: "Custom commission rates for specific categories.",
    defaultValue: DEFAULT_COMMISSION_OVERRIDES
  },
  "platform.feature_flags": {
    key: "platform.feature_flags",
    label: "Feature flags",
    description: "Controlled rollout switches for platform features.",
    defaultValue: DEFAULT_FEATURE_FLAGS
  },
  "platform.maintenance_mode": {
    key: "platform.maintenance_mode",
    label: "Maintenance mode",
    description: "Temporary platform maintenance banner and access behavior.",
    defaultValue: DEFAULT_MAINTENANCE_MODE
  }
};
