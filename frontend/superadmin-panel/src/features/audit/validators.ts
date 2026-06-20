import { DEFAULT_AUDIT_PAGE_SIZE } from "./constants";
import type { AuditLogFilters } from "./types";

export const MIN_AUDIT_EXPORT_REASON_LENGTH = 10;
export const MAX_AUDIT_PAGE_SIZE = 100;

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

function isValidDateTime(value?: string): boolean {
  if (!value?.trim()) {
    return true;
  }

  return !Number.isNaN(new Date(value).getTime());
}

export function validateAuditExportReason(reason: string): ValidationResult {
  const trimmedReason = reason.trim();

  return result(
    trimmedReason.length >= MIN_AUDIT_EXPORT_REASON_LENGTH
      ? []
      : [`Export reason must be at least ${MIN_AUDIT_EXPORT_REASON_LENGTH} characters.`]
  );
}

export function validateAuditLogFilters(filters: AuditLogFilters): ValidationResult {
  const errors: string[] = [];
  const fromTime = filters.from ? new Date(filters.from).getTime() : null;
  const toTime = filters.to ? new Date(filters.to).getTime() : null;

  if (!Number.isInteger(filters.page) || filters.page < 1) {
    errors.push("Page must be at least 1.");
  }

  if (
    !Number.isInteger(filters.page_size) ||
    filters.page_size < DEFAULT_AUDIT_PAGE_SIZE ||
    filters.page_size > MAX_AUDIT_PAGE_SIZE
  ) {
    errors.push(`Page size must be between ${DEFAULT_AUDIT_PAGE_SIZE} and ${MAX_AUDIT_PAGE_SIZE}.`);
  }

  if (!isValidDateTime(filters.from)) {
    errors.push("From date must be valid.");
  }

  if (!isValidDateTime(filters.to)) {
    errors.push("To date must be valid.");
  }

  if (
    fromTime !== null &&
    toTime !== null &&
    !Number.isNaN(fromTime) &&
    !Number.isNaN(toTime) &&
    fromTime > toTime
  ) {
    errors.push("From date must be before to date.");
  }

  return result(errors);
}

function toDateTimeLocalInput(date: Date): string {
  const offsetMs = date.getTimezoneOffset() * 60_000;

  return new Date(date.getTime() - offsetMs).toISOString().slice(0, 16);
}

export function createInitialAuditLogFilters(now = new Date()): AuditLogFilters {
  return {
    page: 1,
    page_size: DEFAULT_AUDIT_PAGE_SIZE,
    from: toDateTimeLocalInput(new Date(now.getTime() - 24 * 60 * 60 * 1000)),
    to: toDateTimeLocalInput(now)
  };
}
