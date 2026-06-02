import {
  DEFAULT_COMMISSION_OVERRIDES,
  DEFAULT_COMMISSION_RATE,
  DEFAULT_FEATURE_FLAGS,
  DEFAULT_MAINTENANCE_MODE
} from "./setting-definitions";
import type {
  CommissionCategoryOverride,
  CommissionCategoryOverridesValue,
  CommissionDefaultRate,
  FeatureFlagsValue,
  MaintenanceModeValue,
  PlatformSetting
} from "./types";

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function readNumber(value: unknown, fallback: number): number {
  return typeof value === "number" && Number.isFinite(value) ? value : fallback;
}

function readString(value: unknown, fallback = ""): string {
  return typeof value === "string" ? value : fallback;
}

function readNullableString(value: unknown): string | null {
  return typeof value === "string" && value.trim() ? value : null;
}

function readBoolean(value: unknown, fallback: boolean): boolean {
  return typeof value === "boolean" ? value : fallback;
}

export function readCommissionDefaultRate(
  setting?: PlatformSetting<unknown>
): CommissionDefaultRate {
  const value = isRecord(setting?.value) ? setting.value : {};

  return {
    rate_percent: readNumber(value.rate_percent, DEFAULT_COMMISSION_RATE.rate_percent),
    applies_to: "all_sellers"
  };
}

function readCommissionOverride(value: unknown): CommissionCategoryOverride | null {
  if (!isRecord(value)) {
    return null;
  }

  const categoryId = readString(value.category_id).trim();

  if (!categoryId) {
    return null;
  }

  return {
    category_id: categoryId,
    rate_percent: readNumber(value.rate_percent, 0)
  };
}

export function readCommissionCategoryOverrides(
  setting?: PlatformSetting<unknown>
): CommissionCategoryOverridesValue {
  const value = isRecord(setting?.value) ? setting.value : {};
  const overrides = Array.isArray(value.overrides)
    ? value.overrides.map(readCommissionOverride).filter((item): item is CommissionCategoryOverride => item !== null)
    : DEFAULT_COMMISSION_OVERRIDES.overrides;

  return { overrides };
}

export function readFeatureFlags(setting?: PlatformSetting<unknown>): FeatureFlagsValue {
  const value = isRecord(setting?.value) ? setting.value : {};
  const rawFlags = isRecord(value.flags) ? value.flags : DEFAULT_FEATURE_FLAGS.flags;
  const flags = Object.entries(rawFlags).reduce<Record<string, boolean>>((accumulator, [key, enabled]) => {
    if (typeof enabled === "boolean") {
      accumulator[key] = enabled;
    }

    return accumulator;
  }, {});

  return { flags };
}

export function readMaintenanceMode(setting?: PlatformSetting<unknown>): MaintenanceModeValue {
  const value = isRecord(setting?.value) ? setting.value : {};

  return {
    enabled: readBoolean(value.enabled, DEFAULT_MAINTENANCE_MODE.enabled),
    message: readString(value.message, DEFAULT_MAINTENANCE_MODE.message),
    starts_at: readNullableString(value.starts_at),
    ends_at: readNullableString(value.ends_at),
    allow_admin_bypass: readBoolean(
      value.allow_admin_bypass,
      DEFAULT_MAINTENANCE_MODE.allow_admin_bypass
    )
  };
}

export function settingConfigured(setting?: PlatformSetting<unknown>): boolean {
  return Boolean(setting);
}
