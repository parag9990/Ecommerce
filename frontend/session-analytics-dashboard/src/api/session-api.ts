import { validateDateRange, type DateRange } from "../lib/date-range";
import {
  ApiError,
  getBlob,
  getJSON,
  sendJSON,
  type BlobResponse
} from "../lib/http";
import { logger } from "../lib/logger";
import {
  channelOptions,
  deviceTypeOptions,
  sourceOptions,
  userTypeOptions,
  type Channel,
  type DeviceType,
  type LiveMetricsResponse,
  type SegmentFilters,
  type TrafficSource,
  type UserType
} from "../features/shell/types";

export type LiveMetricsRequest = {
  dateRange: DateRange;
  filters: SegmentFilters;
};

export const funnelStepKeys = [
  "product_view",
  "add_to_cart",
  "checkout_started",
  "paid"
] as const;

export type FunnelStepKey = (typeof funnelStepKeys)[number];

export type FunnelReportRequest = {
  dateRange: DateRange;
  filters: SegmentFilters;
  steps?: FunnelStepKey[];
};

export const retentionIntervalOptions = [
  { label: "Daily", value: "day" },
  { label: "Weekly", value: "week" },
  { label: "Monthly", value: "month" }
] as const;

export type RetentionInterval = (typeof retentionIntervalOptions)[number]["value"];

export const retentionWindowOptionsByInterval = {
  day: [7, 14, 30],
  month: [3, 6, 12],
  week: [4, 8, 12]
} as const satisfies Record<RetentionInterval, readonly number[]>;

export type RetentionReportRequest = {
  dateRange: DateRange;
  filters: SegmentFilters;
  interval: RetentionInterval;
  window: number;
};

export const analyticsReportTypeOptions = [
  { label: "Overview metrics", value: "overview" },
  { label: "Active sessions", value: "active_sessions" },
  { label: "Journey summaries", value: "journey_summary" },
  { label: "Funnel report", value: "funnel" },
  { label: "Heatmap aggregates", value: "heatmap" },
  { label: "Retention cohorts", value: "retention" }
] as const;

export type AnalyticsReportType =
  (typeof analyticsReportTypeOptions)[number]["value"];

export const reportFormatOptions = [{ label: "CSV", value: "csv" }] as const;

export type ReportFormat = (typeof reportFormatOptions)[number]["value"];

export const reportFrequencyOptions = [
  { label: "Daily", value: "daily" },
  { label: "Weekly", value: "weekly" },
  { label: "Monthly", value: "monthly" }
] as const;

export type ReportFrequency = (typeof reportFrequencyOptions)[number]["value"];

export const reportScheduleStatusOptions = [
  { label: "Active", value: "active" },
  { label: "Paused", value: "paused" },
  { label: "Failed", value: "failed" }
] as const;

export type ReportScheduleStatus =
  (typeof reportScheduleStatusOptions)[number]["value"];

const reportScheduleWritableStatusOptions = [
  { label: "Active", value: "active" },
  { label: "Paused", value: "paused" }
] as const;

export const reportDayOfWeekOptions = [
  { label: "Monday", value: "monday" },
  { label: "Tuesday", value: "tuesday" },
  { label: "Wednesday", value: "wednesday" },
  { label: "Thursday", value: "thursday" },
  { label: "Friday", value: "friday" },
  { label: "Saturday", value: "saturday" },
  { label: "Sunday", value: "sunday" }
] as const;

export type ReportDayOfWeek = (typeof reportDayOfWeekOptions)[number]["value"];

export type ReportFilters = {
  from: string;
  to: string;
  timezone: string;
  channel?: Channel;
  country?: string;
  deviceType?: DeviceType;
  source?: TrafficSource;
  userType?: UserType;
};

export type ExportReportRequest = ReportFilters & {
  format: ReportFormat;
  reportType: AnalyticsReportType;
};

export type ReportSchedule = {
  id: string;
  name: string;
  reportType: AnalyticsReportType;
  format: ReportFormat;
  frequency: ReportFrequency;
  timezone: string;
  timeOfDay: string;
  dayOfWeek?: ReportDayOfWeek;
  dayOfMonth?: number;
  status: ReportScheduleStatus;
  recipients: string[];
  filters: Partial<ReportFilters>;
  lastRunAt?: string;
  nextRunAt?: string;
  lastError?: string;
  createdAt: string;
  updatedAt?: string;
};

export type ReportSchedulesResponse = {
  items: ReportSchedule[];
};

export type CreateReportScheduleInput = {
  name: string;
  reportType: AnalyticsReportType;
  format: ReportFormat;
  frequency: ReportFrequency;
  timezone: string;
  timeOfDay: string;
  dayOfWeek?: ReportDayOfWeek;
  dayOfMonth?: number;
  recipients: string[];
  filters: Partial<ReportFilters>;
};

export type UpdateReportScheduleStatusInput = {
  id: string;
  status: Exclude<ReportScheduleStatus, "failed">;
};

export type RetentionSummary = {
  newUsers: number;
  returningUsers: number;
  returningRate: number;
  averageRetention: number;
  bestCohort?: string;
  worstCohort?: string;
};

export type NewReturningBucket = {
  bucket: string;
  label?: string;
  newUsers: number;
  returningUsers: number;
  totalUsers: number;
};

export type CohortBucket = {
  offset: number;
  label: string;
  users: number;
  rate: number;
  suppressed: boolean;
};

export type RetentionCohort = {
  cohortKey: string;
  cohortLabel: string;
  cohortSize: number;
  buckets: CohortBucket[];
  suppressed: boolean;
};

export type RetentionReportMeta = {
  interval: RetentionInterval;
  window: number;
  from: string;
  to: string;
  generatedAt?: string;
  partial: boolean;
  smallCountThreshold?: number;
  suppressed: boolean;
};

export type RetentionReportResponse = {
  summary: RetentionSummary;
  newVsReturning: NewReturningBucket[];
  cohorts: RetentionCohort[];
  meta: RetentionReportMeta;
};

export type HeatmapDeviceType = "desktop" | "tablet" | "mobile";

export type HeatmapMode = "click" | "scroll";

export const heatmapDeviceTypeOptions = [
  { label: "Desktop", value: "desktop" },
  { label: "Tablet", value: "tablet" },
  { label: "Mobile", value: "mobile" }
] as const satisfies ReadonlyArray<{
  label: string;
  value: HeatmapDeviceType;
}>;

export const heatmapModeOptions = [
  { label: "Clicks", value: "click" },
  { label: "Scroll", value: "scroll" }
] as const satisfies ReadonlyArray<{
  label: string;
  value: HeatmapMode;
}>;

export type HeatmapRequest = {
  path: string;
  deviceType: HeatmapDeviceType;
  from: string;
  to: string;
  mode?: HeatmapMode;
};

export type HeatmapPoint = {
  x: number;
  y: number;
  weight: number;
};

export type HeatmapResponse = {
  points: HeatmapPoint[];
  totalEvents?: number;
  maxWeight?: number;
  averageScrollDepth?: number;
  generatedAt?: string;
  minBucketSize?: number;
  partial?: boolean;
  suppressed?: boolean;
};

export type RawFunnelStep = {
  key?: string;
  step?: string;
  label?: string;
  count?: number;
  sessions?: number;
  uniqueSessions?: number;
  users?: number;
  uniqueUsers?: number;
};

export type RawFunnelReportResponse = {
  steps: RawFunnelStep[];
  generatedAt?: string;
  minSegmentSize?: number;
  partial?: boolean;
  suppressed?: boolean;
};

export type ActiveSessionDeviceType = "desktop" | "mobile" | "tablet" | "unknown";

export const journeyEventTypes = [
  "page_view",
  "product_view",
  "search",
  "click",
  "scroll",
  "add_to_cart",
  "checkout_step",
  "payment_result"
] as const;

export type JourneyEventType = (typeof journeyEventTypes)[number];

export type JourneySessionStatus = "active" | "ended" | "expired" | "unknown";

export const activeSessionDeviceTypeOptions = [
  { label: "All devices", value: "all" },
  { label: "Desktop", value: "desktop" },
  { label: "Mobile", value: "mobile" },
  { label: "Tablet", value: "tablet" },
  { label: "Unknown", value: "unknown" }
] as const satisfies ReadonlyArray<{
  label: string;
  value: ActiveSessionDeviceType | "all";
}>;

export type ActiveSessionDevice = {
  type: ActiveSessionDeviceType;
  browser?: string;
  os?: string;
  userAgent?: string;
};

export type ActiveSessionLocation = {
  country?: string;
  region?: string;
  city?: string;
};

export type ActiveSession = {
  sessionId: string;
  anonymousId: string;
  maskedUserId?: string;
  startedAt: string;
  lastSeenAt: string;
  durationSeconds: number;
  entryPage: string;
  currentPage?: string;
  eventCount: number;
  device: ActiveSessionDevice;
  location: ActiveSessionLocation;
  channel?: string;
  source?: string;
};

export type ActiveSessionBreakdownItem = {
  label: string;
  count: number;
  percentage: number;
};

export type ActiveSessionsRequest = {
  from?: string;
  to?: string;
  q?: string;
  deviceType?: ActiveSessionDeviceType | "all";
  country?: string;
  entryPage?: string;
  limit?: number;
};

export type ActiveSessionsResponse = {
  activeUsers: number;
  activeSessions: number;
  eventsPerMinute: number;
  refreshedAt: string;
  sessions: ActiveSession[];
  deviceBreakdown: ActiveSessionBreakdownItem[];
  locationBreakdown: ActiveSessionBreakdownItem[];
  entryPageBreakdown: ActiveSessionBreakdownItem[];
};

export type JourneySession = {
  sessionId: string;
  anonymousId?: string;
  userId?: string;
  status: JourneySessionStatus;
  entryPage: string;
  exitPage?: string;
  startedAt: string;
  lastSeenAt: string;
  endedAt?: string;
  durationSeconds: number;
  device: ActiveSessionDevice;
  geo?: ActiveSessionLocation;
};

export type JourneySummary = {
  totalEvents: number;
  pageViews: number;
  clicks: number;
  cartActions: number;
  checkoutStarted: boolean;
  paymentCompleted: boolean;
};

export type JourneyEventProperties = Record<string, unknown>;

export type JourneyEvent = {
  eventId: string;
  eventType: JourneyEventType;
  path: string;
  occurredAt: string;
  properties: JourneyEventProperties;
};

export type JourneyResponse = {
  session: JourneySession;
  summary: JourneySummary;
  events: JourneyEvent[];
};

export const maskingModeOptions = [
  { label: "Hidden", value: "hidden" },
  { label: "Masked", value: "masked" },
  { label: "Full", value: "full" }
] as const;

export type MaskingMode = (typeof maskingModeOptions)[number]["value"];

export const locationGranularityOptions = [
  { label: "Hidden", value: "none" },
  { label: "Country only", value: "country" },
  { label: "City", value: "city" }
] as const;

export type LocationGranularity =
  (typeof locationGranularityOptions)[number]["value"];

export type PrivacyMaskingSettings = {
  userIdMode: MaskingMode;
  anonymousIdMode: MaskingMode;
  sessionIdMode: MaskingMode;
  locationGranularity: LocationGranularity;
  showSearchQueries: boolean;
  showIpHash: boolean;
};

export type PrivacyPermissions = {
  canUpdateMasking: boolean;
  canRequestDeletion: boolean;
  canUpdateRetention: boolean;
};

export type PrivacySettingsResponse = {
  masking: PrivacyMaskingSettings;
  permissions: PrivacyPermissions;
  updatedAt: string;
  updatedBy: string;
};

export type RetentionSettings = {
  rawEventsDays: number;
  journeySummariesDays: number;
  heatmapAggregatesDays: number;
  analyticsAggregatesMonths: number;
  activeSessionTtlMinutes: number;
  deletionRequestLogDays: number;
};

export const deletionTargetTypeOptions = [
  { label: "User ID", value: "user_id" },
  { label: "Anonymous ID", value: "anonymous_id" },
  { label: "Session ID", value: "session_id" }
] as const;

export type DeletionTargetType =
  (typeof deletionTargetTypeOptions)[number]["value"];

export type DeletionPreviewRequest = {
  targetType: DeletionTargetType;
  targetValue: string;
};

export type DeletionPreviewResponse = {
  targetType: DeletionTargetType;
  targetValueMasked: string;
  matchedSessions: number;
  matchedEvents: number;
  matchedJourneySummaries: number;
  matchedActiveSessions: number;
  aggregateImpact: "unchanged" | "anonymized" | "aggregates_anonymized_or_unchanged";
  estimatedCompletionSeconds: number;
};

export type CreateDeletionRequest = DeletionPreviewRequest & {
  reason: string;
  confirmed: boolean;
};

export type DeletionRequestStatus =
  | "queued"
  | "processing"
  | "completed"
  | "failed";

export type DeletionRequest = {
  requestId: string;
  targetType: DeletionTargetType;
  targetValueMasked: string;
  status: DeletionRequestStatus;
  requestedBy: string;
  reason: string;
  matchedSessions: number;
  matchedEvents: number;
  matchedJourneySummaries: number;
  matchedActiveSessions: number;
  createdAt: string;
  completedAt?: string;
  error?: string;
};

export type DeletionRequestsResponse = {
  items: DeletionRequest[];
};

type RequestOptions = {
  signal?: AbortSignal;
};

export async function getLiveMetrics(
  request: LiveMetricsRequest,
  options: RequestOptions = {}
): Promise<LiveMetricsResponse> {
  validateLiveMetricsRequest(request);

  const params = new URLSearchParams({
    channel: request.filters.channel,
    device_type: request.filters.deviceType,
    from: request.dateRange.from,
    source: request.filters.source,
    to: request.dateRange.to,
    user_type: request.filters.userType
  });

  try {
    const payload = await getJSON<unknown>(
      `/api/v1/analytics/live?${params.toString()}`,
      { signal: options.signal }
    );

    return parseLiveMetricsResponse(payload);
  } catch (error) {
    logger.warn("analytics.live_metrics.request_failed", {
      code: error instanceof ApiError ? error.code : "UNKNOWN_ERROR",
      status: error instanceof ApiError ? error.status : undefined
    });
    throw error;
  }
}

export async function getFunnelReport(
  request: FunnelReportRequest,
  options: RequestOptions = {}
): Promise<RawFunnelReportResponse> {
  validateFunnelReportRequest(request);

  const params = new URLSearchParams({
    channel: request.filters.channel,
    device_type: request.filters.deviceType,
    from: request.dateRange.from,
    source: request.filters.source,
    to: request.dateRange.to,
    user_type: request.filters.userType
  });

  if (request.steps?.length) {
    params.set("steps", request.steps.join(","));
  }

  try {
    const payload = await getJSON<unknown>(
      `/api/v1/analytics/funnels?${params.toString()}`,
      { signal: options.signal }
    );

    return parseFunnelReportResponse(payload);
  } catch (error) {
    logger.warn("analytics.funnel_report.request_failed", {
      code: error instanceof ApiError ? error.code : "UNKNOWN_ERROR",
      status: error instanceof ApiError ? error.status : undefined
    });
    throw error;
  }
}

export async function getRetentionReport(
  request: RetentionReportRequest,
  options: RequestOptions = {}
): Promise<RetentionReportResponse> {
  validateRetentionReportRequest(request);

  const params = new URLSearchParams({
    from: request.dateRange.from,
    interval: request.interval,
    to: request.dateRange.to,
    window: String(request.window)
  });

  appendQueryParam(params, "device_type", request.filters.deviceType);
  appendQueryParam(params, "channel", request.filters.channel);
  appendQueryParam(params, "source", request.filters.source);
  appendQueryParam(params, "user_type", request.filters.userType);

  try {
    const payload = await getJSON<unknown>(
      `/api/v1/analytics/retention?${params.toString()}`,
      { signal: options.signal }
    );

    return parseRetentionReportResponse(payload, request);
  } catch (error) {
    logger.warn("analytics.retention_report.request_failed", {
      code: error instanceof ApiError ? error.code : "UNKNOWN_ERROR",
      status: error instanceof ApiError ? error.status : undefined
    });
    throw error;
  }
}

export async function getHeatmap(
  request: HeatmapRequest,
  options: RequestOptions = {}
): Promise<HeatmapResponse> {
  const path = normalizeHeatmapPath(request.path);
  validateHeatmapRequest({ ...request, path });

  const params = new URLSearchParams({
    device_type: request.deviceType,
    from: request.from,
    path,
    to: request.to
  });
  appendQueryParam(params, "mode", request.mode);

  try {
    const payload = await getJSON<unknown>(
      `/api/v1/analytics/heatmaps?${params.toString()}`,
      { signal: options.signal }
    );

    return parseHeatmapResponse(payload);
  } catch (error) {
    logger.warn("analytics.heatmap.request_failed", {
      code: error instanceof ApiError ? error.code : "UNKNOWN_ERROR",
      cause:
        error instanceof ApiError && error.cause instanceof Error
          ? error.cause.message
          : undefined,
      message: error instanceof Error ? error.message : "unknown error",
      status: error instanceof ApiError ? error.status : undefined
    });
    throw error;
  }
}

export async function getActiveSessions(
  request: ActiveSessionsRequest,
  options: RequestOptions = {}
): Promise<ActiveSessionsResponse> {
  validateActiveSessionsRequest(request);

  const params = new URLSearchParams();
  params.set("status", "active");
  appendQueryParam(params, "from", request.from);
  appendQueryParam(params, "to", request.to);
  appendQueryParam(params, "q", request.q?.trim());
  appendQueryParam(params, "device_type", request.deviceType);
  appendQueryParam(params, "country", request.country?.trim());
  appendQueryParam(params, "entry_page", request.entryPage?.trim());
  appendQueryParam(params, "limit", request.limit ?? 50);

  try {
    const payload = await getJSON<unknown>(
      `/api/v1/analytics/sessions?${params.toString()}`,
      { signal: options.signal }
    );

    return parseActiveSessionsResponse(payload);
  } catch (error) {
    logger.warn("analytics.active_sessions.request_failed", {
      code: error instanceof ApiError ? error.code : "UNKNOWN_ERROR",
      status: error instanceof ApiError ? error.status : undefined
    });
    throw error;
  }
}

export async function getSessionJourney(
  sessionId: string,
  options: RequestOptions = {}
): Promise<JourneyResponse> {
  const safeSessionId = validateSessionId(sessionId);

  try {
    const payload = await getJSON<unknown>(
      `/api/v1/analytics/sessions/${encodeURIComponent(safeSessionId)}/journey`,
      { signal: options.signal }
    );

    return parseJourneyResponse(payload);
  } catch (error) {
    logger.warn("analytics.session_journey.request_failed", {
      code: error instanceof ApiError ? error.code : "UNKNOWN_ERROR",
      status: error instanceof ApiError ? error.status : undefined
    });
    throw error;
  }
}

export async function exportAnalyticsReport(
  request: ExportReportRequest,
  options: RequestOptions = {}
): Promise<BlobResponse> {
  validateExportReportRequest(request);

  const params = buildReportQuery(request);

  try {
    return await getBlob(`/api/v1/analytics/reports/export?${params.toString()}`, {
      headers: {
        Accept: "text/csv"
      },
      signal: options.signal
    });
  } catch (error) {
    logger.warn("analytics.report_export.request_failed", {
      code: error instanceof ApiError ? error.code : "UNKNOWN_ERROR",
      status: error instanceof ApiError ? error.status : undefined
    });
    throw error;
  }
}

export async function getReportSchedules(
  options: RequestOptions = {}
): Promise<ReportSchedulesResponse> {
  try {
    const payload = await getJSON<unknown>(
      "/api/v1/analytics/reports/schedules",
      { signal: options.signal }
    );

    return parseReportSchedulesResponse(payload);
  } catch (error) {
    logger.warn("analytics.report_schedules.request_failed", {
      code: error instanceof ApiError ? error.code : "UNKNOWN_ERROR",
      status: error instanceof ApiError ? error.status : undefined
    });
    throw error;
  }
}

export async function createReportSchedule(
  input: CreateReportScheduleInput,
  options: RequestOptions = {}
): Promise<ReportSchedule> {
  validateCreateReportScheduleInput(input);

  try {
    const payload = await sendJSON<unknown>(
      "/api/v1/analytics/reports/schedules",
      {
        body: serializeReportScheduleInput(input),
        method: "POST",
        signal: options.signal
      }
    );

    return parseReportSchedule(payload);
  } catch (error) {
    logger.warn("analytics.report_schedule_create.request_failed", {
      code: error instanceof ApiError ? error.code : "UNKNOWN_ERROR",
      status: error instanceof ApiError ? error.status : undefined
    });
    throw error;
  }
}

export async function updateReportScheduleStatus(
  input: UpdateReportScheduleStatusInput,
  options: RequestOptions = {}
): Promise<ReportSchedule> {
  const scheduleId = validateReportScheduleId(input.id);
  assertAllowedValue("status", input.status, reportScheduleWritableStatusOptions);

  try {
    const payload = await sendJSON<unknown>(
      `/api/v1/analytics/reports/schedules/${encodeURIComponent(scheduleId)}`,
      {
        body: { status: input.status },
        method: "PATCH",
        signal: options.signal
      }
    );

    return parseReportSchedule(payload);
  } catch (error) {
    logger.warn("analytics.report_schedule_update.request_failed", {
      code: error instanceof ApiError ? error.code : "UNKNOWN_ERROR",
      status: error instanceof ApiError ? error.status : undefined
    });
    throw error;
  }
}

export async function deleteReportSchedule(
  id: string,
  options: RequestOptions = {}
): Promise<void> {
  const scheduleId = validateReportScheduleId(id);

  try {
    await sendJSON<void>(
      `/api/v1/analytics/reports/schedules/${encodeURIComponent(scheduleId)}`,
      {
        expectEmpty: true,
        method: "DELETE",
        signal: options.signal
      }
    );
  } catch (error) {
    logger.warn("analytics.report_schedule_delete.request_failed", {
      code: error instanceof ApiError ? error.code : "UNKNOWN_ERROR",
      status: error instanceof ApiError ? error.status : undefined
    });
    throw error;
  }
}

export async function getPrivacySettings(
  options: RequestOptions = {}
): Promise<PrivacySettingsResponse> {
  try {
    const payload = await getJSON<unknown>(
      "/api/v1/analytics/privacy/settings",
      { signal: options.signal }
    );

    return parsePrivacySettingsResponse(payload);
  } catch (error) {
    logger.warn("analytics.privacy_settings.request_failed", {
      code: error instanceof ApiError ? error.code : "UNKNOWN_ERROR",
      status: error instanceof ApiError ? error.status : undefined
    });
    throw error;
  }
}

export async function updatePrivacySettings(
  masking: PrivacyMaskingSettings,
  options: RequestOptions = {}
): Promise<PrivacySettingsResponse> {
  validatePrivacyMaskingSettings(masking);

  try {
    const payload = await sendJSON<unknown>(
      "/api/v1/analytics/privacy/settings",
      {
        body: { masking },
        method: "PATCH",
        signal: options.signal
      }
    );

    return parsePrivacySettingsResponse(payload);
  } catch (error) {
    logger.warn("analytics.privacy_settings_update.request_failed", {
      code: error instanceof ApiError ? error.code : "UNKNOWN_ERROR",
      status: error instanceof ApiError ? error.status : undefined
    });
    throw error;
  }
}

export async function getRetentionSettings(
  options: RequestOptions = {}
): Promise<RetentionSettings> {
  try {
    const payload = await getJSON<unknown>(
      "/api/v1/analytics/privacy/retention",
      { signal: options.signal }
    );

    return parseRetentionSettings(payload);
  } catch (error) {
    logger.warn("analytics.privacy_retention.request_failed", {
      code: error instanceof ApiError ? error.code : "UNKNOWN_ERROR",
      status: error instanceof ApiError ? error.status : undefined
    });
    throw error;
  }
}

export async function updateRetentionSettings(
  settings: RetentionSettings,
  reason: string,
  options: RequestOptions = {}
): Promise<RetentionSettings> {
  validatePrivacyRetentionSettings(settings);
  validateAuditReason(reason);

  try {
    const payload = await sendJSON<unknown>(
      "/api/v1/analytics/privacy/retention",
      {
        body: { ...settings, reason: reason.trim() },
        method: "PATCH",
        signal: options.signal
      }
    );

    return parseRetentionSettings(payload);
  } catch (error) {
    logger.warn("analytics.privacy_retention_update.request_failed", {
      code: error instanceof ApiError ? error.code : "UNKNOWN_ERROR",
      status: error instanceof ApiError ? error.status : undefined
    });
    throw error;
  }
}

export async function previewDeletion(
  request: DeletionPreviewRequest,
  options: RequestOptions = {}
): Promise<DeletionPreviewResponse> {
  const safeRequest = validateDeletionPreviewRequest(request);

  try {
    const payload = await sendJSON<unknown>(
      "/api/v1/analytics/privacy/deletion-preview",
      {
        body: safeRequest,
        method: "POST",
        signal: options.signal
      }
    );

    return parseDeletionPreview(payload);
  } catch (error) {
    logger.warn("analytics.privacy_deletion_preview.request_failed", {
      code: error instanceof ApiError ? error.code : "UNKNOWN_ERROR",
      status: error instanceof ApiError ? error.status : undefined
    });
    throw error;
  }
}

export async function createDeletionRequest(
  request: CreateDeletionRequest,
  options: RequestOptions = {}
): Promise<DeletionRequest> {
  const safeRequest = validateCreateDeletionRequest(request);

  try {
    const payload = await sendJSON<unknown>(
      "/api/v1/analytics/privacy/deletion-requests",
      {
        body: safeRequest,
        method: "POST",
        signal: options.signal
      }
    );

    return parseDeletionRequest(payload);
  } catch (error) {
    logger.warn("analytics.privacy_deletion_create.request_failed", {
      code: error instanceof ApiError ? error.code : "UNKNOWN_ERROR",
      status: error instanceof ApiError ? error.status : undefined
    });
    throw error;
  }
}

export async function listDeletionRequests(
  options: RequestOptions = {}
): Promise<DeletionRequestsResponse> {
  try {
    const payload = await getJSON<unknown>(
      "/api/v1/analytics/privacy/deletion-requests",
      { signal: options.signal }
    );

    return parseDeletionRequestsResponse(payload);
  } catch (error) {
    logger.warn("analytics.privacy_deletion_list.request_failed", {
      code: error instanceof ApiError ? error.code : "UNKNOWN_ERROR",
      status: error instanceof ApiError ? error.status : undefined
    });
    throw error;
  }
}

export function validateLiveMetricsRequest(request: LiveMetricsRequest): void {
  const dateValidation = validateDateRange(request.dateRange);
  if (!dateValidation.ok) {
    throw new ApiError({
      code: "VALIDATION_ERROR",
      message: dateValidation.message,
      status: 400
    });
  }

  assertAllowedValue("deviceType", request.filters.deviceType, deviceTypeOptions);
  assertAllowedValue("channel", request.filters.channel, channelOptions);
  assertAllowedValue("source", request.filters.source, sourceOptions);
  assertAllowedValue("userType", request.filters.userType, userTypeOptions);
}

export function validateFunnelReportRequest(request: FunnelReportRequest): void {
  const dateValidation = validateDateRange(request.dateRange);
  if (!dateValidation.ok) {
    throw new ApiError({
      code: "VALIDATION_ERROR",
      message: dateValidation.message,
      status: 400
    });
  }

  assertAllowedValue("deviceType", request.filters.deviceType, deviceTypeOptions);
  assertAllowedValue("channel", request.filters.channel, channelOptions);
  assertAllowedValue("source", request.filters.source, sourceOptions);
  assertAllowedValue("userType", request.filters.userType, userTypeOptions);

  if (request.steps !== undefined && request.steps.length === 0) {
    throw new ApiError({
      code: "VALIDATION_ERROR",
      message: "At least one funnel step is required.",
      status: 400
    });
  }

  for (const step of request.steps ?? []) {
    if (!funnelStepKeys.includes(step)) {
      throw new ApiError({
        code: "VALIDATION_ERROR",
        message: "Invalid funnel step.",
        status: 400
      });
    }
  }
}

export function validateRetentionReportRequest(
  request: RetentionReportRequest
): void {
  const dateValidation = validateDateRange(request.dateRange);
  if (!dateValidation.ok) {
    throw new ApiError({
      code: "VALIDATION_ERROR",
      message: dateValidation.message,
      status: 400
    });
  }

  assertAllowedValue("deviceType", request.filters.deviceType, deviceTypeOptions);
  assertAllowedValue("channel", request.filters.channel, channelOptions);
  assertAllowedValue("source", request.filters.source, sourceOptions);
  assertAllowedValue("userType", request.filters.userType, userTypeOptions);
  assertAllowedValue("interval", request.interval, retentionIntervalOptions);

  const allowedWindows = retentionWindowOptionsByInterval[request.interval];
  if (
    !Number.isInteger(request.window) ||
    !(allowedWindows as readonly number[]).includes(request.window)
  ) {
    throw new ApiError({
      code: "VALIDATION_ERROR",
      message: "Invalid retention window for the selected interval.",
      status: 400
    });
  }
}

export function validateHeatmapRequest(request: HeatmapRequest): void {
  const dateValidation = validateDateRange({
    from: request.from,
    preset: "custom",
    to: request.to
  });
  if (!dateValidation.ok) {
    throw new ApiError({
      code: "VALIDATION_ERROR",
      message: dateValidation.message,
      status: 400
    });
  }

  normalizeHeatmapPath(request.path);
  assertAllowedValue(
    "deviceType",
    request.deviceType,
    heatmapDeviceTypeOptions
  );

  if (request.mode !== undefined) {
    assertAllowedValue("mode", request.mode, heatmapModeOptions);
  }
}

export function validateActiveSessionsRequest(
  request: ActiveSessionsRequest
): void {
  assertAllowedValue(
    "deviceType",
    request.deviceType ?? "all",
    activeSessionDeviceTypeOptions
  );
  assertOptionalLength("q", request.q, 128);
  assertOptionalLength("country", request.country, 80);
  assertOptionalLength("entryPage", request.entryPage, 512);

  if (
    request.limit !== undefined &&
    (!Number.isInteger(request.limit) || request.limit < 1 || request.limit > 200)
  ) {
    throw new ApiError({
      code: "VALIDATION_ERROR",
      message: "Active session limit must be between 1 and 200.",
      status: 400
    });
  }
}

export function validateSessionId(sessionId: string): string {
  const trimmed = sessionId.trim();

  if (trimmed.length === 0) {
    throw new ApiError({
      code: "VALIDATION_ERROR",
      message: "Session id is required.",
      status: 400
    });
  }

  if (trimmed.length > 160 || !/^[A-Za-z0-9._:-]+$/.test(trimmed)) {
    throw new ApiError({
      code: "VALIDATION_ERROR",
      message: "Session id contains unsupported characters.",
      status: 400
    });
  }

  return trimmed;
}

export function validateExportReportRequest(
  request: ExportReportRequest
): void {
  const dateValidation = validateDateRange({
    from: request.from,
    preset: "custom",
    to: request.to
  });
  if (!dateValidation.ok) {
    throw new ApiError({
      code: "VALIDATION_ERROR",
      message: dateValidation.message,
      status: 400
    });
  }

  assertAllowedValue(
    "reportType",
    request.reportType,
    analyticsReportTypeOptions
  );
  assertAllowedValue("format", request.format, reportFormatOptions);
  validateTimezone(request.timezone);
  validateReportFilterValues(request);
}

export function validateCreateReportScheduleInput(
  input: CreateReportScheduleInput
): void {
  validateScheduleName(input.name);
  assertAllowedValue(
    "reportType",
    input.reportType,
    analyticsReportTypeOptions
  );
  assertAllowedValue("format", input.format, reportFormatOptions);
  assertAllowedValue("frequency", input.frequency, reportFrequencyOptions);
  validateTimezone(input.timezone);
  validateTimeOfDay(input.timeOfDay);
  validateScheduleCadence(input);
  validateRecipients(input.recipients);
  validateReportFilterValues(input.filters);
}

export function validatePrivacyMaskingSettings(
  settings: PrivacyMaskingSettings
): void {
  assertAllowedValue("userIdMode", settings.userIdMode, maskingModeOptions);
  assertAllowedValue(
    "anonymousIdMode",
    settings.anonymousIdMode,
    maskingModeOptions
  );
  assertAllowedValue("sessionIdMode", settings.sessionIdMode, maskingModeOptions);
  assertAllowedValue(
    "locationGranularity",
    settings.locationGranularity,
    locationGranularityOptions
  );
}

export function validatePrivacyRetentionSettings(
  settings: RetentionSettings
): void {
  assertIntegerRange("rawEventsDays", settings.rawEventsDays, 7, 180);
  assertIntegerRange(
    "journeySummariesDays",
    settings.journeySummariesDays,
    30,
    730
  );
  assertIntegerRange(
    "heatmapAggregatesDays",
    settings.heatmapAggregatesDays,
    30,
    730
  );
  assertIntegerRange(
    "analyticsAggregatesMonths",
    settings.analyticsAggregatesMonths,
    12,
    84
  );
  assertIntegerRange(
    "activeSessionTtlMinutes",
    settings.activeSessionTtlMinutes,
    15,
    180
  );
  assertIntegerRange(
    "deletionRequestLogDays",
    settings.deletionRequestLogDays,
    365,
    2555
  );
}

function validateDeletionPreviewRequest(
  request: DeletionPreviewRequest
): DeletionPreviewRequest {
  assertAllowedValue("targetType", request.targetType, deletionTargetTypeOptions);
  const targetValue = normalizeDeletionTargetValue(request.targetValue);

  return {
    targetType: request.targetType,
    targetValue
  };
}

function validateCreateDeletionRequest(
  request: CreateDeletionRequest
): CreateDeletionRequest {
  const previewRequest = validateDeletionPreviewRequest(request);
  validateAuditReason(request.reason);

  if (!request.confirmed) {
    throw new ApiError({
      code: "VALIDATION_ERROR",
      message: "Deletion confirmation is required.",
      status: 400
    });
  }

  return {
    ...previewRequest,
    confirmed: true,
    reason: request.reason.trim()
  };
}

function validateAuditReason(reason: string): void {
  const trimmed = reason.trim();

  if (trimmed.length < 10 || trimmed.length > 512) {
    throw new ApiError({
      code: "VALIDATION_ERROR",
      message: "Audit reason must be between 10 and 512 characters.",
      status: 400
    });
  }
}

function normalizeDeletionTargetValue(value: string): string {
  const trimmed = value.trim();

  if (
    trimmed.length < 3 ||
    trimmed.length > 160 ||
    !/^[A-Za-z0-9._:@-]+$/.test(trimmed)
  ) {
    throw new ApiError({
      code: "VALIDATION_ERROR",
      message: "Deletion target contains unsupported characters.",
      status: 400
    });
  }

  return trimmed;
}

function buildReportQuery(request: ExportReportRequest): URLSearchParams {
  const params = new URLSearchParams();

  params.set("report_type", request.reportType);
  params.set("format", request.format);
  params.set("from", request.from);
  params.set("to", request.to);
  params.set("timezone", request.timezone);

  appendQueryParam(params, "device", request.deviceType);
  appendQueryParam(params, "channel", request.channel);
  appendQueryParam(params, "source", request.source);
  appendQueryParam(params, "country", request.country?.trim());
  appendQueryParam(params, "user_type", request.userType);

  return params;
}

function serializeReportScheduleInput(input: CreateReportScheduleInput) {
  return {
    day_of_month:
      input.frequency === "monthly" ? input.dayOfMonth : undefined,
    day_of_week: input.frequency === "weekly" ? input.dayOfWeek : undefined,
    filters: serializeReportFilters(input.filters),
    format: input.format,
    frequency: input.frequency,
    name: input.name.trim(),
    recipients: normalizeRecipients(input.recipients),
    report_type: input.reportType,
    time_of_day: input.timeOfDay,
    timezone: input.timezone
  };
}

function serializeReportFilters(filters: Partial<ReportFilters>) {
  const payload: Record<string, string | number> = {};

  appendSerializedFilter(payload, "from", filters.from);
  appendSerializedFilter(payload, "to", filters.to);
  appendSerializedFilter(payload, "timezone", filters.timezone);
  appendSerializedFilter(payload, "device", filters.deviceType);
  appendSerializedFilter(payload, "channel", filters.channel);
  appendSerializedFilter(payload, "source", filters.source);
  appendSerializedFilter(payload, "country", filters.country?.trim());
  appendSerializedFilter(payload, "user_type", filters.userType);

  return payload;
}

function appendSerializedFilter(
  payload: Record<string, string | number>,
  key: string,
  value?: string | number
) {
  if (value === undefined || value === null || value === "" || value === "all") {
    return;
  }

  payload[key] = value;
}

function parsePrivacySettingsResponse(
  payload: unknown
): PrivacySettingsResponse {
  if (!isRecord(payload)) {
    throw invalidPrivacyResponse();
  }

  const masking = readPrivacyRecord(payload, ["masking"]);
  const permissions = readPrivacyOptionalRecord(payload, ["permissions"]) ?? {};

  return {
    masking: parsePrivacyMaskingSettings(masking),
    permissions: {
      canRequestDeletion:
        readPrivacyOptionalBoolean(permissions, [
          "canRequestDeletion",
          "can_request_deletion"
        ]) ?? false,
      canUpdateMasking:
        readPrivacyOptionalBoolean(permissions, [
          "canUpdateMasking",
          "can_update_masking"
        ]) ?? false,
      canUpdateRetention:
        readPrivacyOptionalBoolean(permissions, [
          "canUpdateRetention",
          "can_update_retention"
        ]) ?? false
    },
    updatedAt:
      readPrivacyOptionalString(payload, ["updatedAt", "updated_at"]) ??
      new Date(0).toISOString(),
    updatedBy:
      readPrivacyOptionalString(payload, ["updatedBy", "updated_by"]) ?? "system"
  };
}

function parsePrivacyMaskingSettings(
  payload: Record<string, unknown>
): PrivacyMaskingSettings {
  const userIdMode = normalizeMaskingMode(
    readPrivacyString(payload, ["userIdMode", "user_id_mode"])
  );
  const anonymousIdMode = normalizeMaskingMode(
    readPrivacyString(payload, ["anonymousIdMode", "anonymous_id_mode"])
  );
  const sessionIdMode = normalizeMaskingMode(
    readPrivacyString(payload, ["sessionIdMode", "session_id_mode"])
  );
  const locationGranularity = normalizeLocationGranularity(
    readPrivacyString(payload, [
      "locationGranularity",
      "location_granularity"
    ])
  );

  return {
    anonymousIdMode,
    locationGranularity,
    sessionIdMode,
    showIpHash:
      readPrivacyOptionalBoolean(payload, ["showIpHash", "show_ip_hash"]) ??
      false,
    showSearchQueries:
      readPrivacyOptionalBoolean(payload, [
        "showSearchQueries",
        "show_search_queries"
      ]) ?? false,
    userIdMode
  };
}

function parseRetentionSettings(payload: unknown): RetentionSettings {
  const record = unwrapRetentionRecord(payload);

  return {
    activeSessionTtlMinutes: readPrivacyInteger(record, [
      "activeSessionTtlMinutes",
      "active_session_ttl_minutes"
    ]),
    analyticsAggregatesMonths: readPrivacyInteger(record, [
      "analyticsAggregatesMonths",
      "analytics_aggregates_months"
    ]),
    deletionRequestLogDays: readPrivacyInteger(record, [
      "deletionRequestLogDays",
      "deletion_request_log_days"
    ]),
    heatmapAggregatesDays: readPrivacyInteger(record, [
      "heatmapAggregatesDays",
      "heatmap_aggregates_days"
    ]),
    journeySummariesDays: readPrivacyInteger(record, [
      "journeySummariesDays",
      "journey_summaries_days"
    ]),
    rawEventsDays: readPrivacyInteger(record, [
      "rawEventsDays",
      "raw_events_days"
    ])
  };
}

function parseDeletionPreview(payload: unknown): DeletionPreviewResponse {
  if (!isRecord(payload)) {
    throw invalidPrivacyResponse();
  }

  return {
    aggregateImpact: normalizeAggregateImpact(
      readPrivacyOptionalString(payload, [
        "aggregateImpact",
        "aggregate_impact"
      ]) ?? "aggregates_anonymized_or_unchanged"
    ),
    estimatedCompletionSeconds: readPrivacyInteger(payload, [
      "estimatedCompletionSeconds",
      "estimated_completion_seconds"
    ]),
    matchedActiveSessions: readPrivacyInteger(payload, [
      "matchedActiveSessions",
      "matched_active_sessions"
    ]),
    matchedEvents: readPrivacyInteger(payload, [
      "matchedEvents",
      "matched_events"
    ]),
    matchedJourneySummaries: readPrivacyInteger(payload, [
      "matchedJourneySummaries",
      "matched_journey_summaries"
    ]),
    matchedSessions: readPrivacyInteger(payload, [
      "matchedSessions",
      "matched_sessions"
    ]),
    targetType: normalizeDeletionTargetType(
      readPrivacyString(payload, ["targetType", "target_type"])
    ),
    targetValueMasked: readPrivacyString(payload, [
      "targetValueMasked",
      "target_value_masked"
    ])
  };
}

function parseDeletionRequestsResponse(
  payload: unknown
): DeletionRequestsResponse {
  if (Array.isArray(payload)) {
    return { items: payload.map(parseDeletionRequest) };
  }

  if (!isRecord(payload)) {
    throw invalidPrivacyResponse();
  }

  const items = readPrivacyOptionalArray(payload, ["items", "requests"]).map(
    parseDeletionRequest
  );
  return { items };
}

function parseDeletionRequest(payload: unknown): DeletionRequest {
  if (!isRecord(payload)) {
    throw invalidPrivacyResponse();
  }

  return {
    completedAt: readPrivacyOptionalString(payload, [
      "completedAt",
      "completed_at"
    ]),
    createdAt: readPrivacyString(payload, ["createdAt", "created_at"]),
    error: readPrivacyOptionalString(payload, ["error"]),
    matchedActiveSessions: readPrivacyInteger(payload, [
      "matchedActiveSessions",
      "matched_active_sessions"
    ]),
    matchedEvents: readPrivacyInteger(payload, [
      "matchedEvents",
      "matched_events"
    ]),
    matchedJourneySummaries: readPrivacyInteger(payload, [
      "matchedJourneySummaries",
      "matched_journey_summaries"
    ]),
    matchedSessions: readPrivacyInteger(payload, [
      "matchedSessions",
      "matched_sessions"
    ]),
    reason: readPrivacyString(payload, ["reason"]),
    requestedBy: readPrivacyString(payload, ["requestedBy", "requested_by"]),
    requestId: readPrivacyString(payload, [
      "requestId",
      "request_id",
      "id",
      "_id"
    ]),
    status: normalizeDeletionRequestStatus(
      readPrivacyString(payload, ["status"])
    ),
    targetType: normalizeDeletionTargetType(
      readPrivacyString(payload, ["targetType", "target_type"])
    ),
    targetValueMasked: readPrivacyString(payload, [
      "targetValueMasked",
      "target_value_masked"
    ])
  };
}

function parseReportSchedulesResponse(payload: unknown): ReportSchedulesResponse {
  if (Array.isArray(payload)) {
    return {
      items: payload.map(parseReportSchedule)
    };
  }

  if (!isRecord(payload)) {
    throw invalidReportScheduleResponse();
  }

  return {
    items: readReportOptionalArray(payload, ["items", "schedules"]).map(
      parseReportSchedule
    )
  };
}

function parseReportSchedule(payload: unknown): ReportSchedule {
  if (!isRecord(payload)) {
    throw invalidReportScheduleResponse();
  }

  const reportType = normalizeReportType(
    readReportString(payload, ["reportType", "report_type"])
  );
  const format = normalizeReportFormat(
    readReportOptionalString(payload, ["format"]) ?? "csv"
  );
  const frequency = normalizeReportFrequency(
    readReportString(payload, ["frequency"])
  );
  const status = normalizeReportScheduleStatus(
    readReportOptionalString(payload, ["status"]) ?? "active"
  );
  const filters = readReportOptionalRecord(payload, ["filters"]);

  return {
    createdAt: readReportString(payload, ["createdAt", "created_at"]),
    dayOfMonth: readReportOptionalPositiveInteger(payload, [
      "dayOfMonth",
      "day_of_month"
    ]),
    dayOfWeek: normalizeOptionalReportDayOfWeek(
      readReportOptionalString(payload, ["dayOfWeek", "day_of_week"])
    ),
    filters: parseReportFilters(filters),
    format,
    frequency,
    id: readReportString(payload, ["id", "scheduleId", "schedule_id", "_id"]),
    lastError: readReportOptionalString(payload, ["lastError", "last_error"]),
    lastRunAt: readReportOptionalString(payload, ["lastRunAt", "last_run_at"]),
    name: readReportString(payload, ["name"]),
    nextRunAt: readReportOptionalString(payload, ["nextRunAt", "next_run_at"]),
    recipients: readReportStringArray(payload, ["recipients"]),
    reportType,
    status,
    timeOfDay: readReportString(payload, ["timeOfDay", "time_of_day"]),
    timezone: readReportString(payload, ["timezone"]),
    updatedAt: readReportOptionalString(payload, ["updatedAt", "updated_at"])
  };
}

function parseReportFilters(
  payload: Record<string, unknown> | undefined
): Partial<ReportFilters> {
  if (!payload) {
    return {};
  }

  return {
    channel: normalizeOptionalChannel(readReportOptionalString(payload, ["channel"])),
    country: readReportOptionalString(payload, ["country"]),
    deviceType: normalizeOptionalDeviceType(
      readReportOptionalString(payload, ["deviceType", "device_type", "device"])
    ),
    from: readReportOptionalString(payload, ["from"]),
    source: normalizeOptionalSource(readReportOptionalString(payload, ["source"])),
    timezone: readReportOptionalString(payload, ["timezone"]),
    to: readReportOptionalString(payload, ["to"]),
    userType: normalizeOptionalUserType(
      readReportOptionalString(payload, ["userType", "user_type"])
    )
  };
}

function validateReportFilterValues(filters: Partial<ReportFilters>): void {
  if (filters.deviceType !== undefined) {
    assertAllowedValue("deviceType", filters.deviceType, deviceTypeOptions);
  }

  if (filters.channel !== undefined) {
    assertAllowedValue("channel", filters.channel, channelOptions);
  }

  if (filters.source !== undefined) {
    assertAllowedValue("source", filters.source, sourceOptions);
  }

  if (filters.userType !== undefined) {
    assertAllowedValue("userType", filters.userType, userTypeOptions);
  }

  assertOptionalLength("country", filters.country, 80);

  if (filters.from !== undefined || filters.to !== undefined) {
    const dateValidation = validateDateRange({
      from: filters.from ?? "",
      preset: "custom",
      to: filters.to ?? ""
    });

    if (!dateValidation.ok) {
      throw new ApiError({
        code: "VALIDATION_ERROR",
        message: dateValidation.message,
        status: 400
      });
    }
  }

  if (filters.timezone !== undefined) {
    validateTimezone(filters.timezone);
  }
}

function validateScheduleCadence(input: CreateReportScheduleInput): void {
  if (input.frequency === "weekly") {
    if (!input.dayOfWeek) {
      throw new ApiError({
        code: "VALIDATION_ERROR",
        message: "Weekly schedules require a day of week.",
        status: 400
      });
    }

    assertAllowedValue("dayOfWeek", input.dayOfWeek, reportDayOfWeekOptions);
  }

  if (input.frequency === "monthly") {
    const dayOfMonth = input.dayOfMonth;
    if (
      dayOfMonth === undefined ||
      !Number.isInteger(dayOfMonth) ||
      dayOfMonth < 1 ||
      dayOfMonth > 31
    ) {
      throw new ApiError({
        code: "VALIDATION_ERROR",
        message: "Monthly schedules require a day between 1 and 31.",
        status: 400
      });
    }
  }
}

function validateScheduleName(name: string): void {
  const trimmed = name.trim();

  if (
    trimmed.length === 0 ||
    trimmed.length > 120 ||
    /[\u0000-\u001f\u007f]/.test(trimmed)
  ) {
    throw new ApiError({
      code: "VALIDATION_ERROR",
      message: "Schedule name must be 1 to 120 readable characters.",
      status: 400
    });
  }
}

function validateTimeOfDay(timeOfDay: string): void {
  if (!/^([01]\d|2[0-3]):[0-5]\d$/.test(timeOfDay)) {
    throw new ApiError({
      code: "VALIDATION_ERROR",
      message: "Schedule time must use HH:mm format.",
      status: 400
    });
  }
}

function validateTimezone(timezone: string): void {
  try {
    new Intl.DateTimeFormat("en-US", { timeZone: timezone }).format();
  } catch {
    throw new ApiError({
      code: "VALIDATION_ERROR",
      message: "Timezone is not supported.",
      status: 400
    });
  }
}

function validateRecipients(recipients: string[]): void {
  const normalized = normalizeRecipients(recipients);

  if (normalized.length === 0 || normalized.length > 10) {
    throw new ApiError({
      code: "VALIDATION_ERROR",
      message: "Add between 1 and 10 report recipients.",
      status: 400
    });
  }

  if (new Set(normalized).size !== normalized.length) {
    throw new ApiError({
      code: "VALIDATION_ERROR",
      message: "Report recipients must be unique.",
      status: 400
    });
  }

  for (const recipient of normalized) {
    if (!isValidEmail(recipient)) {
      throw new ApiError({
        code: "VALIDATION_ERROR",
        message: "Report recipient email is invalid.",
        status: 400
      });
    }
  }
}

function normalizeRecipients(recipients: string[]): string[] {
  return recipients
    .map((recipient) => recipient.trim().toLowerCase())
    .filter(Boolean);
}

function isValidEmail(value: string): boolean {
  return (
    value.length <= 254 &&
    /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value) &&
    !/[\u0000-\u001f\u007f]/.test(value)
  );
}

function validateReportScheduleId(id: string): string {
  const trimmed = id.trim();

  if (trimmed.length === 0 || trimmed.length > 160 || !/^[A-Za-z0-9._:-]+$/.test(trimmed)) {
    throw new ApiError({
      code: "VALIDATION_ERROR",
      message: "Report schedule id contains unsupported characters.",
      status: 400
    });
  }

  return trimmed;
}

function normalizeReportType(value: string): AnalyticsReportType {
  if (
    analyticsReportTypeOptions.some((option) => option.value === value)
  ) {
    return value as AnalyticsReportType;
  }

  throw invalidReportScheduleResponse();
}

function normalizeReportFormat(value: string): ReportFormat {
  if (reportFormatOptions.some((option) => option.value === value)) {
    return value as ReportFormat;
  }

  throw invalidReportScheduleResponse();
}

function normalizeReportFrequency(value: string): ReportFrequency {
  if (reportFrequencyOptions.some((option) => option.value === value)) {
    return value as ReportFrequency;
  }

  throw invalidReportScheduleResponse();
}

function normalizeReportScheduleStatus(value: string): ReportScheduleStatus {
  if (reportScheduleStatusOptions.some((option) => option.value === value)) {
    return value as ReportScheduleStatus;
  }

  throw invalidReportScheduleResponse();
}

function normalizeOptionalReportDayOfWeek(
  value: string | undefined
): ReportDayOfWeek | undefined {
  if (value === undefined) {
    return undefined;
  }

  if (reportDayOfWeekOptions.some((option) => option.value === value)) {
    return value as ReportDayOfWeek;
  }

  throw invalidReportScheduleResponse();
}

function normalizeOptionalDeviceType(
  value: string | undefined
): DeviceType | undefined {
  if (value === undefined) {
    return undefined;
  }

  if (deviceTypeOptions.some((option) => option.value === value)) {
    return value as DeviceType;
  }

  throw invalidReportScheduleResponse();
}

function normalizeOptionalChannel(value: string | undefined): Channel | undefined {
  if (value === undefined) {
    return undefined;
  }

  if (channelOptions.some((option) => option.value === value)) {
    return value as Channel;
  }

  throw invalidReportScheduleResponse();
}

function normalizeOptionalSource(
  value: string | undefined
): TrafficSource | undefined {
  if (value === undefined) {
    return undefined;
  }

  if (sourceOptions.some((option) => option.value === value)) {
    return value as TrafficSource;
  }

  throw invalidReportScheduleResponse();
}

function normalizeOptionalUserType(
  value: string | undefined
): UserType | undefined {
  if (value === undefined) {
    return undefined;
  }

  if (userTypeOptions.some((option) => option.value === value)) {
    return value as UserType;
  }

  throw invalidReportScheduleResponse();
}

function readReportString(
  record: Record<string, unknown>,
  keys: string[]
): string {
  const value = readValue(record, keys);
  if (typeof value !== "string" || value.trim() === "") {
    throw invalidReportScheduleResponse();
  }

  return value;
}

function readReportOptionalString(
  record: Record<string, unknown>,
  keys: string[]
): string | undefined {
  const value = readValue(record, keys);
  if (value === undefined || value === null || value === "") {
    return undefined;
  }

  if (typeof value !== "string") {
    throw invalidReportScheduleResponse();
  }

  return value;
}

function readReportOptionalPositiveInteger(
  record: Record<string, unknown>,
  keys: string[]
): number | undefined {
  const value = readValue(record, keys);
  if (value === undefined || value === null || value === "") {
    return undefined;
  }

  const parsed = typeof value === "number" ? value : Number(value);
  if (!Number.isInteger(parsed) || parsed < 1) {
    throw invalidReportScheduleResponse();
  }

  return parsed;
}

function readReportStringArray(
  record: Record<string, unknown>,
  keys: string[]
): string[] {
  const value = readValue(record, keys);
  if (!Array.isArray(value)) {
    throw invalidReportScheduleResponse();
  }

  return value.map((item) => {
    if (typeof item !== "string" || item.trim() === "") {
      throw invalidReportScheduleResponse();
    }

    return item;
  });
}

function readReportOptionalArray(
  record: Record<string, unknown>,
  keys: string[]
): unknown[] {
  const value = readValue(record, keys);
  if (value === undefined || value === null) {
    return [];
  }

  if (!Array.isArray(value)) {
    throw invalidReportScheduleResponse();
  }

  return value;
}

function readReportOptionalRecord(
  record: Record<string, unknown>,
  keys: string[]
): Record<string, unknown> | undefined {
  const value = readValue(record, keys);
  if (value === undefined || value === null) {
    return undefined;
  }

  if (!isRecord(value)) {
    throw invalidReportScheduleResponse();
  }

  return value;
}

function parseLiveMetricsResponse(payload: unknown): LiveMetricsResponse {
  if (!isRecord(payload)) {
    throw invalidMetricsResponse();
  }

  return {
    activeUsersNow: readNumber(payload, "activeUsersNow"),
    averageSessionDurationSeconds: readNumber(
      payload,
      "averageSessionDurationSeconds"
    ),
    bounceRate: readNumber(payload, "bounceRate"),
    conversionRate: readNumber(payload, "conversionRate"),
    productViewToCartRate: readNumber(payload, "productViewToCartRate"),
    sessionsToday: readNumber(payload, "sessionsToday")
  };
}

function parseFunnelReportResponse(payload: unknown): RawFunnelReportResponse {
  if (!isRecord(payload)) {
    throw invalidFunnelResponse();
  }

  return {
    generatedAt: readFunnelOptionalString(payload, ["generatedAt", "generated_at"]),
    minSegmentSize: readFunnelOptionalNumber(payload, [
      "minSegmentSize",
      "min_segment_size"
    ]),
    partial: readFunnelOptionalBoolean(payload, ["partial", "is_partial"]),
    steps: readFunnelOptionalArray(payload, ["steps"]).map(parseRawFunnelStep),
    suppressed: readFunnelOptionalBoolean(payload, ["suppressed", "is_suppressed"])
  };
}

function parseRetentionReportResponse(
  payload: unknown,
  request: RetentionReportRequest
): RetentionReportResponse {
  if (!isRecord(payload)) {
    throw invalidRetentionResponse();
  }

  const summaryPayload = readRetentionOptionalRecord(payload, ["summary"]);
  const metaPayload = readRetentionOptionalRecord(payload, ["meta"]);

  return {
    cohorts: readRetentionOptionalArray(payload, ["cohorts"]).map(
      parseRetentionCohort
    ),
    meta: parseRetentionMeta(metaPayload, request),
    newVsReturning: readRetentionOptionalArray(payload, [
      "newVsReturning",
      "new_vs_returning"
    ]).map(parseNewReturningBucket),
    summary: parseRetentionSummary(summaryPayload)
  };
}

function parseRetentionSummary(
  payload: Record<string, unknown> | undefined
): RetentionSummary {
  if (!payload) {
    return {
      averageRetention: 0,
      newUsers: 0,
      returningRate: 0,
      returningUsers: 0
    };
  }

  return {
    averageRetention: sanitizeRetentionRate(
      readRetentionOptionalNumber(payload, [
        "averageRetention",
        "average_retention"
      ]) ?? 0
    ),
    bestCohort: readRetentionOptionalString(payload, [
      "bestCohort",
      "best_cohort"
    ]),
    newUsers: sanitizeRetentionCount(
      readRetentionOptionalNumber(payload, ["newUsers", "new_users"]) ?? 0
    ),
    returningRate: sanitizeRetentionRate(
      readRetentionOptionalNumber(payload, [
        "returningRate",
        "returning_rate"
      ]) ?? 0
    ),
    returningUsers: sanitizeRetentionCount(
      readRetentionOptionalNumber(payload, [
        "returningUsers",
        "returning_users"
      ]) ?? 0
    ),
    worstCohort: readRetentionOptionalString(payload, [
      "worstCohort",
      "worst_cohort"
    ])
  };
}

function parseNewReturningBucket(payload: unknown): NewReturningBucket {
  if (!isRecord(payload)) {
    throw invalidRetentionResponse();
  }

  const newUsers = sanitizeRetentionCount(
    readRetentionOptionalNumber(payload, ["newUsers", "new_users"]) ?? 0
  );
  const returningUsers = sanitizeRetentionCount(
    readRetentionOptionalNumber(payload, [
      "returningUsers",
      "returning_users"
    ]) ?? 0
  );

  return {
    bucket: readRetentionString(payload, ["bucket"]),
    label: readRetentionOptionalString(payload, [
      "label",
      "bucketLabel",
      "bucket_label"
    ]),
    newUsers,
    returningUsers,
    totalUsers:
      sanitizeRetentionCount(
        readRetentionOptionalNumber(payload, ["totalUsers", "total_users"]) ??
          newUsers + returningUsers
      )
  };
}

function parseRetentionCohort(payload: unknown): RetentionCohort {
  if (!isRecord(payload)) {
    throw invalidRetentionResponse();
  }

  const cohortKey = readRetentionString(payload, ["cohortKey", "cohort_key"]);

  return {
    buckets: readRetentionOptionalArray(payload, ["buckets"]).map(
      parseRetentionBucket
    ),
    cohortKey,
    cohortLabel:
      readRetentionOptionalString(payload, ["cohortLabel", "cohort_label"]) ??
      cohortKey,
    cohortSize: sanitizeRetentionCount(
      readRetentionOptionalNumber(payload, ["cohortSize", "cohort_size"]) ?? 0
    ),
    suppressed:
      readRetentionOptionalBoolean(payload, ["suppressed", "is_suppressed"]) ??
      false
  };
}

function parseRetentionBucket(payload: unknown): CohortBucket {
  if (!isRecord(payload)) {
    throw invalidRetentionResponse();
  }

  const offset = Math.max(
    0,
    Math.round(readRetentionOptionalNumber(payload, ["offset"]) ?? 0)
  );

  return {
    label:
      readRetentionOptionalString(payload, ["label"]) ??
      `+${offset.toString()}`,
    offset,
    rate: sanitizeRetentionRate(
      readRetentionOptionalNumber(payload, ["rate", "retention_rate"]) ?? 0
    ),
    suppressed:
      readRetentionOptionalBoolean(payload, ["suppressed", "is_suppressed"]) ??
      false,
    users: sanitizeRetentionCount(
      readRetentionOptionalNumber(payload, [
        "users",
        "retainedUsers",
        "retained_users"
      ]) ?? 0
    )
  };
}

function parseRetentionMeta(
  payload: Record<string, unknown> | undefined,
  request: RetentionReportRequest
): RetentionReportMeta {
  const rawInterval = payload
    ? readRetentionOptionalString(payload, ["interval"])
    : undefined;
  const from = payload
    ? readRetentionOptionalString(payload, ["from"])
    : undefined;
  const to = payload
    ? readRetentionOptionalString(payload, ["to"])
    : undefined;
  const window = payload
    ? readRetentionOptionalNumber(payload, ["window"])
    : undefined;
  const interval = retentionIntervalOptions.some(
    (option) => option.value === rawInterval
  )
    ? (rawInterval as RetentionInterval)
    : request.interval;

  return {
    from: from ?? request.dateRange.from,
    generatedAt: payload
      ? readRetentionOptionalString(payload, ["generatedAt", "generated_at"])
      : undefined,
    interval,
    partial: payload
      ? readRetentionOptionalBoolean(payload, ["partial", "is_partial"]) ?? false
      : false,
    smallCountThreshold: payload
      ? readRetentionOptionalNumber(payload, [
          "smallCountThreshold",
          "small_count_threshold",
          "minSegmentSize",
          "min_segment_size"
        ])
      : undefined,
    suppressed: payload
      ? readRetentionOptionalBoolean(payload, ["suppressed", "is_suppressed"]) ??
        false
      : false,
    to: to ?? request.dateRange.to,
    window: window !== undefined ? Math.round(window) : request.window
  };
}

function sanitizeRetentionCount(value: number): number {
  if (!Number.isFinite(value)) {
    return 0;
  }

  return Math.max(0, Math.round(value));
}

function sanitizeRetentionRate(value: number): number {
  if (!Number.isFinite(value)) {
    return 0;
  }

  return Math.min(100, Math.max(0, value));
}

function parseHeatmapResponse(payload: unknown): HeatmapResponse {
  if (!isRecord(payload)) {
    throw invalidHeatmapResponse();
  }

  return {
    averageScrollDepth: readHeatmapOptionalNumber(payload, [
      "averageScrollDepth",
      "average_scroll_depth"
    ]),
    generatedAt: readHeatmapOptionalString(payload, ["generatedAt", "generated_at"]),
    maxWeight: readHeatmapOptionalNumber(payload, ["maxWeight", "max_weight"]),
    minBucketSize: readHeatmapOptionalNumber(payload, [
      "minBucketSize",
      "min_bucket_size"
    ]),
    partial: readHeatmapOptionalBoolean(payload, ["partial", "is_partial"]),
    points: readHeatmapArray(payload, ["points"]).map(parseHeatmapPoint),
    suppressed: readHeatmapOptionalBoolean(payload, ["suppressed", "is_suppressed"]),
    totalEvents: readHeatmapOptionalNumber(payload, ["totalEvents", "total_events"])
  };
}

function parseHeatmapPoint(payload: unknown): HeatmapPoint {
  if (!isRecord(payload)) {
    throw invalidHeatmapResponse();
  }

  const x = readHeatmapNumber(payload, ["x"]);
  const y = readHeatmapNumber(payload, ["y"]);
  const weight = readHeatmapNumber(payload, ["weight"]);

  if (weight < 0) {
    throw invalidHeatmapResponse();
  }

  return { weight, x, y };
}

function parseRawFunnelStep(payload: unknown): RawFunnelStep {
  if (!isRecord(payload)) {
    throw invalidFunnelResponse();
  }

  return {
    count: readFunnelOptionalNumber(payload, ["count"]),
    key: readFunnelOptionalString(payload, ["key", "stepKey", "step_key"]),
    label: readFunnelOptionalString(payload, ["label", "name"]),
    sessions: readFunnelOptionalNumber(payload, ["sessions"]),
    step: readFunnelOptionalString(payload, ["step"]),
    uniqueSessions: readFunnelOptionalNumber(payload, [
      "uniqueSessions",
      "unique_sessions"
    ]),
    uniqueUsers: readFunnelOptionalNumber(payload, ["uniqueUsers", "unique_users"]),
    users: readFunnelOptionalNumber(payload, ["users"])
  };
}

function parseActiveSessionsResponse(payload: unknown): ActiveSessionsResponse {
  if (!isRecord(payload)) {
    throw invalidActiveSessionsResponse();
  }

  const sessionsPayload = readArray(payload, ["sessions"]);

  return {
    activeUsers: readFiniteNumber(payload, ["activeUsers", "active_users"]),
    activeSessions: readFiniteNumber(payload, [
      "activeSessions",
      "active_sessions"
    ]),
    eventsPerMinute: readFiniteNumber(payload, [
      "eventsPerMinute",
      "events_per_minute"
    ]),
    refreshedAt: readString(payload, ["refreshedAt", "refreshed_at"]),
    sessions: sessionsPayload.map(parseActiveSession),
    deviceBreakdown: readBreakdown(payload, [
      "deviceBreakdown",
      "device_breakdown"
    ]),
    locationBreakdown: readBreakdown(payload, [
      "locationBreakdown",
      "location_breakdown"
    ]),
    entryPageBreakdown: readBreakdown(payload, [
      "entryPageBreakdown",
      "entry_page_breakdown"
    ])
  };
}

function parseActiveSession(payload: unknown): ActiveSession {
  if (!isRecord(payload)) {
    throw invalidActiveSessionsResponse();
  }

  const device = readOptionalRecord(payload, ["device"]) ?? {};
  const location = readOptionalRecord(payload, ["location"]) ?? {};
  const deviceType = readOptionalString(device, ["type"]) ?? "unknown";

  return {
    anonymousId: readString(payload, ["anonymousId", "anonymous_id"]),
    channel: readOptionalString(payload, ["channel"]),
    currentPage: readOptionalString(payload, ["currentPage", "current_page"]),
    device: {
      browser: readOptionalString(device, ["browser"]),
      os: readOptionalString(device, ["os"]),
      type: normalizeDeviceType(deviceType),
      userAgent: readOptionalString(device, ["userAgent", "user_agent"])
    },
    durationSeconds: readFiniteNumber(payload, [
      "durationSeconds",
      "duration_seconds"
    ]),
    entryPage: readString(payload, ["entryPage", "entry_page"]),
    eventCount: readFiniteNumber(payload, ["eventCount", "event_count"]),
    lastSeenAt: readString(payload, ["lastSeenAt", "last_seen_at"]),
    location: {
      city: readOptionalString(location, ["city"]),
      country: readOptionalString(location, ["country"]),
      region: readOptionalString(location, ["region"])
    },
    maskedUserId: readOptionalString(payload, ["maskedUserId", "masked_user_id"]),
    sessionId: readString(payload, ["sessionId", "session_id"]),
    source: readOptionalString(payload, ["source"]),
    startedAt: readString(payload, ["startedAt", "started_at"])
  };
}

function parseJourneyResponse(payload: unknown): JourneyResponse {
  if (!isRecord(payload)) {
    throw invalidJourneyResponse();
  }

  const events = readJourneyOptionalArray(payload, ["events"]).map(
    parseJourneyEvent
  );
  const derivedSummary = deriveJourneySummary(events);
  const summaryPayload = readJourneyOptionalRecord(payload, ["summary"]);

  return {
    events,
    session: parseJourneySession(
      readJourneyRecord(payload, ["session"]),
      events
    ),
    summary: summaryPayload
      ? parseJourneySummary(summaryPayload, derivedSummary)
      : derivedSummary
  };
}

function parseJourneySession(
  payload: Record<string, unknown>,
  events: JourneyEvent[]
): JourneySession {
  const device = readJourneyOptionalRecord(payload, ["device"]) ?? {};
  const geo = readJourneyOptionalRecord(payload, ["geo", "location"]);
  const firstEvent = events[0];
  const lastEvent = events[events.length - 1];
  const startedAt =
    readJourneyOptionalString(payload, ["startedAt", "started_at"]) ??
    firstEvent?.occurredAt;

  if (!startedAt) {
    throw invalidJourneyResponse();
  }

  const endedAt = readJourneyOptionalString(payload, ["endedAt", "ended_at"]);
  const lastSeenAt =
    readJourneyOptionalString(payload, ["lastSeenAt", "last_seen_at"]) ??
    endedAt ??
    lastEvent?.occurredAt ??
    startedAt;
  const durationSeconds =
    readJourneyOptionalFiniteNumber(payload, [
      "durationSeconds",
      "duration_seconds"
    ]) ?? calculateDurationSeconds(startedAt, endedAt ?? lastSeenAt);

  return {
    anonymousId: readJourneyOptionalString(payload, [
      "anonymousId",
      "anonymous_id"
    ]),
    device: {
      browser: readJourneyOptionalString(device, ["browser"]),
      os: readJourneyOptionalString(device, ["os"]),
      type: normalizeDeviceType(
        readJourneyOptionalString(device, ["type", "deviceType", "device_type"]) ??
          "unknown"
      ),
      userAgent: readJourneyOptionalString(device, ["userAgent", "user_agent"])
    },
    durationSeconds,
    endedAt,
    entryPage:
      readJourneyOptionalString(payload, ["entryPage", "entry_page"]) ??
      firstEvent?.path ??
      "-",
    exitPage:
      readJourneyOptionalString(payload, ["exitPage", "exit_page"]) ??
      (endedAt ? lastEvent?.path : undefined),
    geo: geo
      ? {
          city: readJourneyOptionalString(geo, ["city"]),
          country: readJourneyOptionalString(geo, ["country"]),
          region: readJourneyOptionalString(geo, ["region"])
        }
      : undefined,
    lastSeenAt,
    sessionId: readJourneyString(payload, ["sessionId", "session_id"]),
    startedAt,
    status: normalizeJourneyStatus(
      readJourneyOptionalString(payload, ["status"]) ?? (endedAt ? "ended" : "active")
    ),
    userId: readJourneyOptionalString(payload, ["userId", "user_id"])
  };
}

function parseJourneySummary(
  payload: Record<string, unknown>,
  fallback: JourneySummary
): JourneySummary {
  return {
    cartActions:
      readJourneyOptionalFiniteNumber(payload, ["cartActions", "cart_actions"]) ??
      fallback.cartActions,
    checkoutStarted:
      readJourneyOptionalBoolean(payload, [
        "checkoutStarted",
        "checkout_started"
      ]) ?? fallback.checkoutStarted,
    clicks:
      readJourneyOptionalFiniteNumber(payload, ["clicks"]) ?? fallback.clicks,
    pageViews:
      readJourneyOptionalFiniteNumber(payload, ["pageViews", "page_views"]) ??
      fallback.pageViews,
    paymentCompleted:
      readJourneyOptionalBoolean(payload, [
        "paymentCompleted",
        "payment_completed"
      ]) ?? fallback.paymentCompleted,
    totalEvents:
      readJourneyOptionalFiniteNumber(payload, ["totalEvents", "total_events"]) ??
      fallback.totalEvents
  };
}

function parseJourneyEvent(payload: unknown, index: number): JourneyEvent {
  if (!isRecord(payload)) {
    throw invalidJourneyResponse();
  }

  const eventType = normalizeJourneyEventType(
    readJourneyString(payload, ["eventType", "event_type"])
  );
  const occurredAt = readJourneyString(payload, ["occurredAt", "occurred_at"]);
  const sessionId =
    readJourneyOptionalString(payload, ["sessionId", "session_id"]) ?? "session";

  return {
    eventId:
      readJourneyOptionalString(payload, ["eventId", "event_id", "_id"]) ??
      `${sessionId}:${occurredAt}:${index}`,
    eventType,
    occurredAt,
    path: readJourneyOptionalString(payload, ["path"]) ?? "-",
    properties: readJourneyProperties(payload)
  };
}

function deriveJourneySummary(events: JourneyEvent[]): JourneySummary {
  return {
    cartActions: events.filter((event) => event.eventType === "add_to_cart")
      .length,
    checkoutStarted: events.some((event) => event.eventType === "checkout_step"),
    clicks: events.filter((event) => event.eventType === "click").length,
    pageViews: events.filter((event) => event.eventType === "page_view").length,
    paymentCompleted: events.some((event) => {
      if (event.eventType !== "payment_result") {
        return false;
      }

      const status = String(event.properties.status ?? "").toLowerCase();
      return ["completed", "paid", "success", "succeeded"].includes(status);
    }),
    totalEvents: events.length
  };
}

function readBreakdown(
  payload: Record<string, unknown>,
  keys: string[]
): ActiveSessionBreakdownItem[] {
  return readArray(payload, keys).map((item) => {
    if (!isRecord(item)) {
      throw invalidActiveSessionsResponse();
    }

    return {
      count: readFiniteNumber(item, ["count"]),
      label: readString(item, ["label"]),
      percentage: readFiniteNumber(item, ["percentage"])
    };
  });
}

function readNumber(record: Record<string, unknown>, key: keyof LiveMetricsResponse) {
  const value = record[key];
  if (typeof value !== "number" || !Number.isFinite(value)) {
    throw invalidMetricsResponse();
  }

  return value;
}

function readFiniteNumber(record: Record<string, unknown>, keys: string[]): number {
  const value = readValue(record, keys);
  if (typeof value !== "number" || !Number.isFinite(value)) {
    throw invalidActiveSessionsResponse();
  }

  return value;
}

function readString(record: Record<string, unknown>, keys: string[]): string {
  const value = readValue(record, keys);
  if (typeof value !== "string" || value.trim() === "") {
    throw invalidActiveSessionsResponse();
  }

  return value;
}

function readOptionalString(
  record: Record<string, unknown>,
  keys: string[]
): string | undefined {
  const value = readValue(record, keys);
  if (value === undefined || value === null || value === "") {
    return undefined;
  }

  if (typeof value !== "string") {
    throw invalidActiveSessionsResponse();
  }

  return value;
}

function readArray(record: Record<string, unknown>, keys: string[]): unknown[] {
  const value = readValue(record, keys);
  if (!Array.isArray(value)) {
    throw invalidActiveSessionsResponse();
  }

  return value;
}

function readOptionalRecord(
  record: Record<string, unknown>,
  keys: string[]
): Record<string, unknown> | undefined {
  const value = readValue(record, keys);
  if (value === undefined || value === null) {
    return undefined;
  }

  if (!isRecord(value)) {
    throw invalidActiveSessionsResponse();
  }

  return value;
}

function readJourneyRecord(
  record: Record<string, unknown>,
  keys: string[]
): Record<string, unknown> {
  const value = readValue(record, keys);
  if (!isRecord(value)) {
    throw invalidJourneyResponse();
  }

  return value;
}

function readJourneyString(record: Record<string, unknown>, keys: string[]): string {
  const value = readValue(record, keys);
  if (typeof value !== "string" || value.trim() === "") {
    throw invalidJourneyResponse();
  }

  return value;
}

function readJourneyOptionalString(
  record: Record<string, unknown>,
  keys: string[]
): string | undefined {
  const value = readValue(record, keys);
  if (value === undefined || value === null || value === "") {
    return undefined;
  }

  if (typeof value !== "string") {
    throw invalidJourneyResponse();
  }

  return value;
}

function readJourneyOptionalBoolean(
  record: Record<string, unknown>,
  keys: string[]
): boolean | undefined {
  const value = readValue(record, keys);
  if (value === undefined || value === null || value === "") {
    return undefined;
  }

  if (typeof value !== "boolean") {
    throw invalidJourneyResponse();
  }

  return value;
}

function readJourneyOptionalFiniteNumber(
  record: Record<string, unknown>,
  keys: string[]
): number | undefined {
  const value = readValue(record, keys);
  if (value === undefined || value === null || value === "") {
    return undefined;
  }

  if (typeof value !== "number" || !Number.isFinite(value)) {
    throw invalidJourneyResponse();
  }

  return value;
}

function readJourneyOptionalArray(
  record: Record<string, unknown>,
  keys: string[]
): unknown[] {
  const value = readValue(record, keys);
  if (value === undefined || value === null) {
    return [];
  }

  if (!Array.isArray(value)) {
    throw invalidJourneyResponse();
  }

  return value;
}

function readFunnelOptionalString(
  record: Record<string, unknown>,
  keys: string[]
): string | undefined {
  const value = readValue(record, keys);
  if (value === undefined || value === null || value === "") {
    return undefined;
  }

  if (typeof value !== "string") {
    throw invalidFunnelResponse();
  }

  return value;
}

function readFunnelOptionalNumber(
  record: Record<string, unknown>,
  keys: string[]
): number | undefined {
  const value = readValue(record, keys);
  if (value === undefined || value === null || value === "") {
    return undefined;
  }

  if (typeof value === "number" && Number.isFinite(value)) {
    return value;
  }

  if (typeof value === "string") {
    const parsed = Number(value);
    if (Number.isFinite(parsed)) {
      return parsed;
    }
  }

  throw invalidFunnelResponse();
}

function readFunnelOptionalBoolean(
  record: Record<string, unknown>,
  keys: string[]
): boolean | undefined {
  const value = readValue(record, keys);
  if (value === undefined || value === null || value === "") {
    return undefined;
  }

  if (typeof value !== "boolean") {
    throw invalidFunnelResponse();
  }

  return value;
}

function readFunnelOptionalArray(
  record: Record<string, unknown>,
  keys: string[]
): unknown[] {
  const value = readValue(record, keys);
  if (value === undefined || value === null) {
    return [];
  }

  if (!Array.isArray(value)) {
    throw invalidFunnelResponse();
  }

  return value;
}

function readRetentionString(
  record: Record<string, unknown>,
  keys: string[]
): string {
  const value = readValue(record, keys);
  if (typeof value !== "string" || value.trim() === "") {
    throw invalidRetentionResponse();
  }

  return value;
}

function readRetentionOptionalString(
  record: Record<string, unknown>,
  keys: string[]
): string | undefined {
  const value = readValue(record, keys);
  if (value === undefined || value === null || value === "") {
    return undefined;
  }

  if (typeof value !== "string") {
    throw invalidRetentionResponse();
  }

  return value;
}

function readRetentionOptionalNumber(
  record: Record<string, unknown>,
  keys: string[]
): number | undefined {
  const value = readValue(record, keys);
  if (value === undefined || value === null || value === "") {
    return undefined;
  }

  if (typeof value === "number" && Number.isFinite(value)) {
    return value;
  }

  if (typeof value === "string") {
    const parsed = Number(value);
    if (Number.isFinite(parsed)) {
      return parsed;
    }
  }

  throw invalidRetentionResponse();
}

function readRetentionOptionalBoolean(
  record: Record<string, unknown>,
  keys: string[]
): boolean | undefined {
  const value = readValue(record, keys);
  if (value === undefined || value === null || value === "") {
    return undefined;
  }

  if (typeof value !== "boolean") {
    throw invalidRetentionResponse();
  }

  return value;
}

function readRetentionOptionalArray(
  record: Record<string, unknown>,
  keys: string[]
): unknown[] {
  const value = readValue(record, keys);
  if (value === undefined || value === null) {
    return [];
  }

  if (!Array.isArray(value)) {
    throw invalidRetentionResponse();
  }

  return value;
}

function readRetentionOptionalRecord(
  record: Record<string, unknown>,
  keys: string[]
): Record<string, unknown> | undefined {
  const value = readValue(record, keys);
  if (value === undefined || value === null) {
    return undefined;
  }

  if (!isRecord(value)) {
    throw invalidRetentionResponse();
  }

  return value;
}

function readHeatmapNumber(record: Record<string, unknown>, keys: string[]): number {
  const value = readValue(record, keys);
  if (typeof value === "number" && Number.isFinite(value)) {
    return value;
  }

  throw invalidHeatmapResponse();
}

function readHeatmapOptionalNumber(
  record: Record<string, unknown>,
  keys: string[]
): number | undefined {
  const value = readValue(record, keys);
  if (value === undefined || value === null || value === "") {
    return undefined;
  }

  if (typeof value === "number" && Number.isFinite(value)) {
    return value;
  }

  if (typeof value === "string") {
    const parsed = Number(value);
    if (Number.isFinite(parsed)) {
      return parsed;
    }
  }

  throw invalidHeatmapResponse();
}

function readHeatmapOptionalBoolean(
  record: Record<string, unknown>,
  keys: string[]
): boolean | undefined {
  const value = readValue(record, keys);
  if (value === undefined || value === null || value === "") {
    return undefined;
  }

  if (typeof value !== "boolean") {
    throw invalidHeatmapResponse();
  }

  return value;
}

function readHeatmapOptionalString(
  record: Record<string, unknown>,
  keys: string[]
): string | undefined {
  const value = readValue(record, keys);
  if (value === undefined || value === null || value === "") {
    return undefined;
  }

  if (typeof value !== "string") {
    throw invalidHeatmapResponse();
  }

  return value;
}

function readHeatmapArray(
  record: Record<string, unknown>,
  keys: string[]
): unknown[] {
  const value = readValue(record, keys);
  if (!Array.isArray(value)) {
    throw invalidHeatmapResponse();
  }

  return value;
}

function readJourneyOptionalRecord(
  record: Record<string, unknown>,
  keys: string[]
): Record<string, unknown> | undefined {
  const value = readValue(record, keys);
  if (value === undefined || value === null) {
    return undefined;
  }

  if (!isRecord(value)) {
    throw invalidJourneyResponse();
  }

  return value;
}

function readJourneyProperties(
  record: Record<string, unknown>
): JourneyEventProperties {
  const value = readValue(record, ["properties"]);
  if (value === undefined || value === null) {
    return {};
  }

  if (!isRecord(value)) {
    throw invalidJourneyResponse();
  }

  return value;
}

function unwrapRetentionRecord(payload: unknown): Record<string, unknown> {
  if (!isRecord(payload)) {
    throw invalidPrivacyResponse();
  }

  const nested = readValue(payload, ["retention"]);
  if (nested === undefined || nested === null) {
    return payload;
  }
  if (!isRecord(nested)) {
    throw invalidPrivacyResponse();
  }
  return nested;
}

function readPrivacyRecord(
  record: Record<string, unknown>,
  keys: string[]
): Record<string, unknown> {
  const value = readValue(record, keys);
  if (!isRecord(value)) {
    throw invalidPrivacyResponse();
  }
  return value;
}

function readPrivacyOptionalRecord(
  record: Record<string, unknown>,
  keys: string[]
): Record<string, unknown> | undefined {
  const value = readValue(record, keys);
  if (value === undefined || value === null) {
    return undefined;
  }
  if (!isRecord(value)) {
    throw invalidPrivacyResponse();
  }
  return value;
}

function readPrivacyString(
  record: Record<string, unknown>,
  keys: string[]
): string {
  const value = readValue(record, keys);
  if (typeof value !== "string" || value.trim() === "") {
    throw invalidPrivacyResponse();
  }
  return value;
}

function readPrivacyOptionalString(
  record: Record<string, unknown>,
  keys: string[]
): string | undefined {
  const value = readValue(record, keys);
  if (value === undefined || value === null || value === "") {
    return undefined;
  }
  if (typeof value !== "string") {
    throw invalidPrivacyResponse();
  }
  return value;
}

function readPrivacyInteger(
  record: Record<string, unknown>,
  keys: string[]
): number {
  const value = readValue(record, keys);
  const parsed = typeof value === "number" ? value : Number(value);

  if (!Number.isInteger(parsed)) {
    throw invalidPrivacyResponse();
  }
  return parsed;
}

function readPrivacyOptionalBoolean(
  record: Record<string, unknown>,
  keys: string[]
): boolean | undefined {
  const value = readValue(record, keys);
  if (value === undefined || value === null || value === "") {
    return undefined;
  }
  if (typeof value !== "boolean") {
    throw invalidPrivacyResponse();
  }
  return value;
}

function readPrivacyOptionalArray(
  record: Record<string, unknown>,
  keys: string[]
): unknown[] {
  const value = readValue(record, keys);
  if (value === undefined || value === null) {
    return [];
  }
  if (!Array.isArray(value)) {
    throw invalidPrivacyResponse();
  }
  return value;
}

function readValue(record: Record<string, unknown>, keys: string[]): unknown {
  for (const key of keys) {
    if (key in record) {
      return record[key];
    }
  }

  return undefined;
}

function assertAllowedValue<T extends string>(
  field: string,
  value: string,
  options: ReadonlyArray<{ value: T }>
) {
  if (!options.some((option) => option.value === value)) {
    throw new ApiError({
      code: "VALIDATION_ERROR",
      message: `Invalid ${field} filter.`,
      status: 400
    });
  }
}

function assertOptionalLength(
  field: string,
  value: string | undefined,
  maxLength: number
) {
  if (value !== undefined && value.length > maxLength) {
    throw new ApiError({
      code: "VALIDATION_ERROR",
      message: `${field} filter is too long.`,
      status: 400
    });
  }
}

function assertIntegerRange(
  field: string,
  value: number,
  min: number,
  max: number
) {
  if (!Number.isInteger(value) || value < min || value > max) {
    throw new ApiError({
      code: "VALIDATION_ERROR",
      message: `${field} must be between ${min} and ${max}.`,
      status: 400
    });
  }
}

function appendQueryParam(
  params: URLSearchParams,
  key: string,
  value?: string | number
) {
  if (value === undefined || value === null || value === "" || value === "all") {
    return;
  }

  params.set(key, String(value));
}

function normalizeHeatmapPath(path: string): string {
  const trimmed = path.trim();

  if (trimmed.length === 0 || trimmed.length > 512 || !trimmed.startsWith("/")) {
    throw new ApiError({
      code: "VALIDATION_ERROR",
      message: "Heatmap path must be a site path.",
      status: 400
    });
  }

  if (/[\u0000-\u001f\u007f]/.test(trimmed)) {
    throw new ApiError({
      code: "VALIDATION_ERROR",
      message: "Heatmap path contains unsupported characters.",
      status: 400
    });
  }

  return trimmed;
}

function normalizeDeviceType(value: string): ActiveSessionDeviceType {
  if (["desktop", "mobile", "tablet", "unknown"].includes(value)) {
    return value as ActiveSessionDeviceType;
  }

  return "unknown";
}

function normalizeJourneyEventType(value: string): JourneyEventType {
  if (journeyEventTypes.includes(value as JourneyEventType)) {
    return value as JourneyEventType;
  }

  throw invalidJourneyResponse();
}

function normalizeJourneyStatus(value: string): JourneySessionStatus {
  if (["active", "ended", "expired"].includes(value)) {
    return value as JourneySessionStatus;
  }

  return "unknown";
}

function normalizeMaskingMode(value: string): MaskingMode {
  if (maskingModeOptions.some((option) => option.value === value)) {
    return value as MaskingMode;
  }

  throw invalidPrivacyResponse();
}

function normalizeLocationGranularity(value: string): LocationGranularity {
  if (locationGranularityOptions.some((option) => option.value === value)) {
    return value as LocationGranularity;
  }

  throw invalidPrivacyResponse();
}

function normalizeDeletionTargetType(value: string): DeletionTargetType {
  if (deletionTargetTypeOptions.some((option) => option.value === value)) {
    return value as DeletionTargetType;
  }

  throw invalidPrivacyResponse();
}

function normalizeDeletionRequestStatus(value: string): DeletionRequestStatus {
  if (["queued", "processing", "completed", "failed"].includes(value)) {
    return value as DeletionRequestStatus;
  }

  throw invalidPrivacyResponse();
}

function normalizeAggregateImpact(
  value: string
): DeletionPreviewResponse["aggregateImpact"] {
  if (
    value === "unchanged" ||
    value === "anonymized" ||
    value === "aggregates_anonymized_or_unchanged"
  ) {
    return value;
  }

  throw invalidPrivacyResponse();
}

function calculateDurationSeconds(startedAt: string, lastSeenAt: string): number {
  const started = new Date(startedAt).getTime();
  const lastSeen = new Date(lastSeenAt).getTime();

  if (Number.isNaN(started) || Number.isNaN(lastSeen) || lastSeen < started) {
    return 0;
  }

  return Math.round((lastSeen - started) / 1000);
}

function invalidMetricsResponse(): ApiError {
  return new ApiError({
    code: "INVALID_ANALYTICS_RESPONSE",
    message: "Analytics service returned an unexpected metrics response."
  });
}

function invalidActiveSessionsResponse(): ApiError {
  return new ApiError({
    code: "INVALID_ACTIVE_SESSIONS_RESPONSE",
    message: "Analytics service returned an unexpected active sessions response."
  });
}

function invalidFunnelResponse(): ApiError {
  return new ApiError({
    code: "INVALID_FUNNEL_RESPONSE",
    message: "Analytics service returned an unexpected funnel response."
  });
}

function invalidRetentionResponse(): ApiError {
  return new ApiError({
    code: "INVALID_RETENTION_RESPONSE",
    message: "Analytics service returned an unexpected retention response."
  });
}

function invalidJourneyResponse(): ApiError {
  return new ApiError({
    code: "INVALID_SESSION_JOURNEY_RESPONSE",
    message: "Analytics service returned an unexpected session journey response."
  });
}

function invalidHeatmapResponse(): ApiError {
  return new ApiError({
    code: "INVALID_HEATMAP_RESPONSE",
    message: "Analytics service returned an unexpected heatmap response."
  });
}

function invalidReportScheduleResponse(): ApiError {
  return new ApiError({
    code: "INVALID_REPORT_SCHEDULE_RESPONSE",
    message: "Analytics service returned an unexpected report schedule response."
  });
}

function invalidPrivacyResponse(): ApiError {
  return new ApiError({
    code: "INVALID_PRIVACY_RESPONSE",
    message: "Analytics service returned an unexpected privacy response."
  });
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}
