import type { ReportFrequency } from "../../../api/session-api";

export function formatReportRunTime(value?: string): string {
  if (!value) {
    return "-";
  }

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return new Intl.DateTimeFormat("en-US", {
    dateStyle: "medium",
    timeStyle: "short"
  }).format(date);
}

export function formatReportFrequency(frequency: ReportFrequency): string {
  return `${frequency.charAt(0).toUpperCase()}${frequency.slice(1)}`;
}
