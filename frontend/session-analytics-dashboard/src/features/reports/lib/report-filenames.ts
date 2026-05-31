import type { AnalyticsReportType } from "../../../api/session-api";

const reportFilePrefix: Record<AnalyticsReportType, string> = {
  active_sessions: "active_sessions",
  funnel: "funnel",
  heatmap: "heatmap",
  journey_summary: "journey_summary",
  overview: "overview",
  retention: "retention"
};

export function buildCsvFilename({
  from,
  reportType,
  to
}: {
  from: string;
  reportType: AnalyticsReportType;
  to: string;
}): string {
  const timestamp = new Date()
    .toISOString()
    .slice(0, 19)
    .replace(/[:T]/g, "-");

  return `${reportFilePrefix[reportType]}_${sanitizeDatePart(from)}_${sanitizeDatePart(
    to
  )}_${timestamp}.csv`;
}

export function sanitizeCsvFilename(filename: string): string {
  const sanitized = filename
    .trim()
    .replace(/[\\/:*?"<>|\u0000-\u001f]/g, "_")
    .replace(/\s+/g, "_")
    .slice(0, 160);

  return sanitized.endsWith(".csv") ? sanitized : `${sanitized || "report"}.csv`;
}

function sanitizeDatePart(value: string): string {
  return value.replace(/[^0-9-]/g, "");
}
