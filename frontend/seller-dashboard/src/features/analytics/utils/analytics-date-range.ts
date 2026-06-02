import type { AnalyticsDateRange, AnalyticsPreset } from "../types";

const MS_PER_DAY = 24 * 60 * 60 * 1000;

function utcDateOnly(date: Date) {
  return new Date(Date.UTC(date.getUTCFullYear(), date.getUTCMonth(), date.getUTCDate()));
}

function startOfUtcDay(date: Date) {
  return utcDateOnly(date).toISOString();
}

function endOfUtcDay(date: Date) {
  const start = utcDateOnly(date);
  return new Date(start.getTime() + MS_PER_DAY - 1).toISOString();
}

function parseDateInput(value: string) {
  if (!/^\d{4}-\d{2}-\d{2}$/.test(value)) {
    return null;
  }

  const [year, month, day] = value.split("-").map(Number);
  const parsed = new Date(Date.UTC(year, month - 1, day));

  if (
    parsed.getUTCFullYear() !== year ||
    parsed.getUTCMonth() !== month - 1 ||
    parsed.getUTCDate() !== day
  ) {
    return null;
  }

  return parsed;
}

export function createPresetRange(
  preset: AnalyticsPreset,
  now = new Date(),
): AnalyticsDateRange {
  const daysByPreset: Record<AnalyticsPreset, number> = {
    "7d": 7,
    "30d": 30,
    "90d": 90,
  };
  const end = utcDateOnly(now);
  const start = new Date(end.getTime() - (daysByPreset[preset] - 1) * MS_PER_DAY);

  return {
    preset,
    from: startOfUtcDay(start),
    to: endOfUtcDay(end),
  };
}

export function createCustomRange(fromDate: string, toDate: string): AnalyticsDateRange {
  const parsedFrom = parseDateInput(fromDate);
  const parsedTo = parseDateInput(toDate);
  const fallback = utcDateOnly(new Date());
  const from = parsedFrom ?? parsedTo ?? fallback;
  const to = parsedTo ?? parsedFrom ?? fallback;
  const [start, end] = from.getTime() <= to.getTime() ? [from, to] : [to, from];

  return {
    preset: "custom",
    from: startOfUtcDay(start),
    to: endOfUtcDay(end),
  };
}

export function toDateInputValue(value: string) {
  const date = new Date(value);

  if (Number.isNaN(date.getTime())) {
    return "";
  }

  return date.toISOString().slice(0, 10);
}

export function formatDateRangeLabel(range: AnalyticsDateRange) {
  const from = toDateInputValue(range.from);
  const to = toDateInputValue(range.to);

  if (!from || !to) {
    return "Selected range";
  }

  return `${from} to ${to}`;
}
