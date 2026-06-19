import { format, subDays } from "date-fns";

export type DatePreset = "today" | "7d" | "30d" | "custom";

export type DateRange = {
  preset: DatePreset;
  from: string;
  to: string;
};

export type DateRangeValidation =
  | { ok: true }
  | { ok: false; message: string };

export const DATE_PRESETS: Array<{ label: string; value: Exclude<DatePreset, "custom"> }> = [
  { label: "Today", value: "today" },
  { label: "7D", value: "7d" },
  { label: "30D", value: "30d" }
];

export function getDefaultDateRange(now: Date = new Date()): DateRange {
  return getDateRangeForPreset("7d", now);
}

export function getDateRangeForPreset(
  preset: Exclude<DatePreset, "custom">,
  now: Date = new Date()
): DateRange {
  const today = toISODate(now);

  if (preset === "today") {
    return {
      preset,
      from: today,
      to: today
    };
  }

  const days = preset === "7d" ? 6 : 29;

  return {
    preset,
    from: toISODate(subDays(now, days)),
    to: today
  };
}

export function toISODate(date: Date): string {
  return format(date, "yyyy-MM-dd");
}

export function validateDateRange(range: DateRange): DateRangeValidation {
  if (!isISODate(range.from) || !isISODate(range.to)) {
    return {
      ok: false,
      message: "Use a valid start and end date."
    };
  }

  if (range.from > range.to) {
    return {
      ok: false,
      message: "Start date must be before or equal to end date."
    };
  }

  return { ok: true };
}

function isISODate(value: string): boolean {
  if (!/^\d{4}-\d{2}-\d{2}$/.test(value)) {
    return false;
  }

  const date = new Date(`${value}T00:00:00.000Z`);
  return !Number.isNaN(date.getTime()) && toISODate(date) === value;
}
