import type {
  CommissionCategoryOverride,
  CommissionCategoryOverridesValue,
  CommissionDefaultRate,
  FeatureFlagsValue,
  MaintenanceModeStatus,
  MaintenanceModeValue,
  SearchSynonymInput
} from "./types";

export const MIN_SETTING_REASON_LENGTH = 10;
export const MIN_MAINTENANCE_MESSAGE_LENGTH = 10;
export const FEATURE_FLAG_KEY_PATTERN = /^[a-z][a-z0-9_]*$/;

export type ValidationResult = {
  valid: boolean;
  errors: string[];
};

function result(errors: string[]): ValidationResult {
  return {
    valid: errors.length === 0,
    errors
  };
}

function isFiniteNumber(value: number): boolean {
  return Number.isFinite(value) && !Number.isNaN(value);
}

function validateRate(rate: number, label: string): string[] {
  if (!isFiniteNumber(rate)) {
    return [`${label} must be a number.`];
  }

  if (rate < 0 || rate > 50) {
    return [`${label} must be between 0 and 50 percent.`];
  }

  return [];
}

export function validateReason(reason: string): ValidationResult {
  const trimmedReason = reason.trim();

  return result(
    trimmedReason.length >= MIN_SETTING_REASON_LENGTH
      ? []
      : [`Reason must be at least ${MIN_SETTING_REASON_LENGTH} characters.`]
  );
}

export function validateCommissionDefaultRate(value: CommissionDefaultRate): ValidationResult {
  const errors = validateRate(value.rate_percent, "Default commission");

  if (value.applies_to !== "all_sellers") {
    errors.push("Default commission can only apply to all sellers.");
  }

  return result(errors);
}

export function validateCommissionCategoryOverrides(
  value: CommissionCategoryOverridesValue
): ValidationResult {
  const errors: string[] = [];
  const seenCategoryIds = new Set<string>();

  value.overrides.forEach((override, index) => {
    const rowLabel = `Override ${index + 1}`;
    const categoryId = override.category_id.trim();

    if (!categoryId) {
      errors.push(`${rowLabel} category id is required.`);
    }

    if (categoryId && seenCategoryIds.has(categoryId)) {
      errors.push(`Duplicate category override for ${categoryId}.`);
    }

    seenCategoryIds.add(categoryId);
    errors.push(...validateRate(override.rate_percent, `${rowLabel} commission`));
  });

  return result(errors);
}

export function validateFeatureFlags(value: FeatureFlagsValue): ValidationResult {
  const errors = Object.keys(value.flags)
    .filter((flagKey) => !FEATURE_FLAG_KEY_PATTERN.test(flagKey))
    .map((flagKey) => `Feature flag "${flagKey}" must use lowercase snake_case.`);

  return result(errors);
}

function parseTimestamp(value: string | null): number | null {
  if (!value) {
    return null;
  }

  const timestamp = new Date(value).getTime();

  return Number.isNaN(timestamp) ? Number.NaN : timestamp;
}

export function validateMaintenanceMode(value: MaintenanceModeValue): ValidationResult {
  const errors: string[] = [];
  const startsAt = parseTimestamp(value.starts_at);
  const endsAt = parseTimestamp(value.ends_at);

  if (value.enabled && value.message.trim().length < MIN_MAINTENANCE_MESSAGE_LENGTH) {
    errors.push(
      `Maintenance message must be at least ${MIN_MAINTENANCE_MESSAGE_LENGTH} characters when enabled.`
    );
  }

  if (Number.isNaN(startsAt)) {
    errors.push("Maintenance start time must be a valid date.");
  }

  if (Number.isNaN(endsAt)) {
    errors.push("Maintenance end time must be a valid date.");
  }

  if (startsAt !== null && endsAt !== null && !Number.isNaN(startsAt) && !Number.isNaN(endsAt)) {
    if (endsAt <= startsAt) {
      errors.push("Maintenance end time must be after the start time.");
    }
  }

  return result(errors);
}

export function getMaintenanceModeStatus(
  value: MaintenanceModeValue,
  now: Date = new Date()
): MaintenanceModeStatus {
  if (!value.enabled) {
    return "disabled";
  }

  const nowTime = now.getTime();
  const startsAt = parseTimestamp(value.starts_at);
  const endsAt = parseTimestamp(value.ends_at);

  if (startsAt !== null && !Number.isNaN(startsAt) && startsAt > nowTime) {
    return "scheduled";
  }

  if (endsAt !== null && !Number.isNaN(endsAt) && endsAt <= nowTime) {
    return "expired";
  }

  return "active";
}

export function normalizeSearchTerm(value: string): string {
  return value.trim().toLowerCase().replace(/\s+/g, " ");
}

export function normalizeSearchSynonymInput(input: SearchSynonymInput): SearchSynonymInput {
  const root = normalizeSearchTerm(input.root);
  const seen = new Set<string>();
  const synonyms = input.synonyms
    .map(normalizeSearchTerm)
    .filter((term) => term.length > 0 && term !== root)
    .filter((term) => {
      if (seen.has(term)) {
        return false;
      }

      seen.add(term);
      return true;
    });

  return { root, synonyms };
}

export function validateSearchSynonymInput(input: SearchSynonymInput): ValidationResult {
  const normalized = normalizeSearchSynonymInput(input);
  const errors: string[] = [];

  if (!normalized.root) {
    errors.push("Root term is required.");
  }

  if (normalized.synonyms.length === 0) {
    errors.push("At least one synonym is required.");
  }

  return result(errors);
}

export function normalizeCommissionOverrides(
  overrides: readonly CommissionCategoryOverride[]
): CommissionCategoryOverridesValue {
  return {
    overrides: overrides
      .map((override) => ({
        category_id: override.category_id.trim(),
        rate_percent: override.rate_percent
      }))
      .filter((override) => override.category_id.length > 0)
  };
}
