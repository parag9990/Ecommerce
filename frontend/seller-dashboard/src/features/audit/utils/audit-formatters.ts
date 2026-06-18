import type { AuditJson } from "../types";

export type ChangedField = {
  field: string;
  before: unknown;
  after: unknown;
};

const relativeFormatter = new Intl.RelativeTimeFormat("en", {
  numeric: "auto",
});

const relativeUnits = [
  { unit: "year", seconds: 365 * 24 * 60 * 60 },
  { unit: "month", seconds: 30 * 24 * 60 * 60 },
  { unit: "week", seconds: 7 * 24 * 60 * 60 },
  { unit: "day", seconds: 24 * 60 * 60 },
  { unit: "hour", seconds: 60 * 60 },
  { unit: "minute", seconds: 60 },
] as const;

const sensitiveFieldPatterns = [
  /password/i,
  /token/i,
  /secret/i,
  /api[_-]?key/i,
  /authorization/i,
  /otp/i,
  /cvv/i,
  /card/i,
  /payment/i,
];

function parseDate(value: string) {
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? null : date;
}

export function formatAuditTime(value: string) {
  const date = parseDate(value);

  if (!date) {
    return "Unknown time";
  }

  return new Intl.DateTimeFormat("en-IN", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}

export function formatRelativeAuditTime(value: string, nowMs = Date.now()) {
  const date = parseDate(value);

  if (!date) {
    return "Unknown time";
  }

  const diffSeconds = Math.round((date.getTime() - nowMs) / 1000);
  const absoluteSeconds = Math.abs(diffSeconds);

  if (absoluteSeconds < 45) {
    return "just now";
  }

  const match =
    relativeUnits.find((candidate) => absoluteSeconds >= candidate.seconds) ??
    relativeUnits[relativeUnits.length - 1];
  const valueInUnit = Math.round(diffSeconds / match.seconds);

  return relativeFormatter.format(valueInUnit, match.unit);
}

export function getChangedFields(before: AuditJson, after: AuditJson): ChangedField[] {
  if (!before || !after) {
    return [];
  }

  const keys = new Set([...Object.keys(before), ...Object.keys(after)]);

  return Array.from(keys)
    .filter((key) => JSON.stringify(before[key]) !== JSON.stringify(after[key]))
    .map((key) => ({
      field: key,
      before: before[key],
      after: after[key],
    }));
}

export function shouldRedactAuditField(field: string) {
  return sensitiveFieldPatterns.some((pattern) => pattern.test(field));
}

export function formatAuditValue(value: unknown, field?: string) {
  if (field && shouldRedactAuditField(field)) {
    return "Redacted";
  }

  if (value === null || value === undefined || value === "") {
    return "Empty";
  }

  if (typeof value === "boolean") {
    return value ? "Yes" : "No";
  }

  if (typeof value === "object") {
    const serialized = JSON.stringify(value);
    return serialized.length > 90 ? `${serialized.slice(0, 87)}...` : serialized;
  }

  return String(value);
}
