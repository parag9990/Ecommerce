import {
  retentionWindowOptionsByInterval,
  type RetentionInterval
} from "../../../api/session-api";

export function formatRetentionCount(value: number): string {
  return new Intl.NumberFormat("en-US", {
    maximumFractionDigits: 1,
    notation: value >= 10000 ? "compact" : "standard"
  }).format(guardNumber(value));
}

export function formatRetentionRate(value: number): string {
  return `${guardNumber(value).toFixed(1)}%`;
}

export function formatOffsetLabel(
  interval: RetentionInterval,
  offset: number
): string {
  if (offset === 0) {
    return interval === "day" ? "D0" : interval === "week" ? "W0" : "M0";
  }

  const prefix = interval === "day" ? "D" : interval === "week" ? "W" : "M";
  return `${prefix}+${offset}`;
}

export function getRetentionWindowOptions(
  interval: RetentionInterval
): readonly number[] {
  return retentionWindowOptionsByInterval[interval];
}

export function formatWindowLabel(
  interval: RetentionInterval,
  window: number
): string {
  const unit =
    interval === "day" ? "days" : interval === "week" ? "weeks" : "months";

  return `${window} ${unit}`;
}

function guardNumber(value: number): number {
  return Number.isFinite(value) ? Math.max(0, value) : 0;
}
