import type { AnalyticsReportType } from "../../../api/session-api";

export const reportLabels: Record<AnalyticsReportType, string> = {
  active_sessions: "Active sessions",
  funnel: "Funnel report",
  heatmap: "Heatmap aggregates",
  journey_summary: "Journey summaries",
  overview: "Overview metrics",
  retention: "Retention cohorts"
};

export const reportDescriptions: Record<AnalyticsReportType, string> = {
  active_sessions: "Current session summaries without raw identifiers.",
  funnel: "Step counts, conversion rates, and drop-off rates.",
  heatmap: "Aggregate click and scroll buckets by page and device.",
  journey_summary: "One row per session journey summary.",
  overview: "High-level dashboard metrics for the selected range.",
  retention: "Cohort size and return behavior by offset."
};
