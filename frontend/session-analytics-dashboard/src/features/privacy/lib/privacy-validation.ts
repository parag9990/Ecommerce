import type { DeletionTargetType, RetentionSettings } from "../../../api/session-api";

export type RetentionErrors = Partial<Record<keyof RetentionSettings, string>>;

export function validateRetentionSettings(
  settings: RetentionSettings
): RetentionErrors {
  return {
    ...rangeError("rawEventsDays", settings.rawEventsDays, 7, 180, "days"),
    ...rangeError(
      "journeySummariesDays",
      settings.journeySummariesDays,
      30,
      730,
      "days"
    ),
    ...rangeError(
      "heatmapAggregatesDays",
      settings.heatmapAggregatesDays,
      30,
      730,
      "days"
    ),
    ...rangeError(
      "analyticsAggregatesMonths",
      settings.analyticsAggregatesMonths,
      12,
      84,
      "months"
    ),
    ...rangeError(
      "activeSessionTtlMinutes",
      settings.activeSessionTtlMinutes,
      15,
      180,
      "minutes"
    ),
    ...rangeError(
      "deletionRequestLogDays",
      settings.deletionRequestLogDays,
      365,
      2555,
      "days"
    )
  };
}

export function validateDeletionReason(reason: string): string | undefined {
  const trimmed = reason.trim();
  if (trimmed.length < 10) {
    return "Add an audit reason of at least 10 characters.";
  }
  if (trimmed.length > 512) {
    return "Audit reason must be 512 characters or fewer.";
  }
  return undefined;
}

export function validateDeletionTargetValue(
  targetType: DeletionTargetType,
  value: string
): string | undefined {
  const trimmed = value.trim();
  if (trimmed.length < 3) {
    return `${targetLabel(targetType)} must be at least 3 characters.`;
  }
  if (trimmed.length > 160 || !/^[A-Za-z0-9._:@-]+$/.test(trimmed)) {
    return `${targetLabel(targetType)} contains unsupported characters.`;
  }
  return undefined;
}

function rangeError(
  key: keyof RetentionSettings,
  value: number,
  min: number,
  max: number,
  unit: string
): RetentionErrors {
  if (!Number.isInteger(value) || value < min || value > max) {
    return {
      [key]: `Use ${min} to ${max} ${unit}.`
    };
  }
  return {};
}

function targetLabel(targetType: DeletionTargetType): string {
  if (targetType === "anonymous_id") {
    return "Anonymous ID";
  }
  if (targetType === "session_id") {
    return "Session ID";
  }
  return "User ID";
}
