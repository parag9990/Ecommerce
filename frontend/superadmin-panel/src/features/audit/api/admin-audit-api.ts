import { apiFetch, apiFetchBlob } from "../../../lib/http";
import type {
  AdminAuditLog,
  AuditExportRequest,
  AuditLogFilters,
  AuditLogListResponse,
  AuditSummary
} from "../types";

const AUDIT_LOGS_PATH = "/api/v1/admin/audit-logs";
const AUDIT_LOG_EXPORT_PATH = "/api/v1/admin/audit-logs/export";

type RawAuditLog = Partial<
  Omit<
    AdminAuditLog,
    "id" | "actor_admin_id" | "resource_id" | "request_id" | "ip_hash" | "created_at"
  >
> & {
  id?: string | number | null;
  actor_admin_id?: string | number | null;
  actor_id?: string | number | null;
  actor_role?: string | null;
  action?: string | null;
  resource_type?: string | null;
  resource_id?: string | number | null;
  request_id?: string | number | null;
  ip_hash?: string | number | null;
  before_summary?: unknown;
  after_summary?: unknown;
  reason?: string | null;
  created_at?: string | null;
  timestamp?: string | null;
};

type RawAuditLogListResponse = {
  logs?: RawAuditLog[] | null;
  audit_logs?: RawAuditLog[] | null;
  items?: RawAuditLog[] | null;
  total?: number | null;
  page?: number | null;
  page_size?: number | null;
  limit?: number | null;
};

function stringValue(value: string | number | null | undefined, fallback: string): string {
  const text = String(value ?? "").trim();

  return text || fallback;
}

function numberValue(value: number | null | undefined): number | undefined {
  return typeof value === "number" && Number.isFinite(value) ? value : undefined;
}

function isObjectSummary(value: unknown): value is AuditSummary {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function normalizeSummary(value: unknown): AuditSummary | null {
  return isObjectSummary(value) ? value : null;
}

function normalizeAuditLog(log: RawAuditLog, index: number): AdminAuditLog {
  const action = stringValue(log.action, "unknown.action");
  const resourceType = stringValue(log.resource_type, "unknown");
  const resourceId = stringValue(log.resource_id, "unknown_resource");
  const createdAt = log.created_at ?? log.timestamp ?? null;

  return {
    id: stringValue(log.id, `audit_log_${index + 1}`),
    actor_admin_id: stringValue(log.actor_admin_id ?? log.actor_id, "unknown_admin"),
    actor_role: stringValue(log.actor_role, "unknown_role"),
    action,
    resource_type: resourceType,
    resource_id: resourceId,
    request_id: stringValue(log.request_id, "unknown_request"),
    ip_hash: stringValue(log.ip_hash, "unavailable"),
    before_summary: normalizeSummary(log.before_summary),
    after_summary: normalizeSummary(log.after_summary),
    reason: log.reason?.trim() || null,
    created_at: createdAt
  };
}

function toApiDateTime(value?: string): string | undefined {
  const trimmed = value?.trim();

  if (!trimmed) {
    return undefined;
  }

  const date = new Date(trimmed);

  return Number.isNaN(date.getTime()) ? trimmed : date.toISOString();
}

function buildQueryString(values: Record<string, string | number | undefined>): string {
  const params = new URLSearchParams();

  Object.entries(values).forEach(([key, value]) => {
    if (value !== undefined && value !== "") {
      params.set(key, String(value));
    }
  });

  return params.toString();
}

function appendQuery(path: string, query: string): string {
  return query ? `${path}?${query}` : path;
}

function toApiFilterPayload(filters: AuditLogFilters): Record<string, string | number | undefined> {
  return {
    page: filters.page,
    page_size: filters.page_size,
    actor_id: filters.actor_id?.trim(),
    action: filters.action?.trim(),
    resource_type:
      filters.resource_type && filters.resource_type !== "all"
        ? filters.resource_type.trim()
        : undefined,
    resource_id: filters.resource_id?.trim(),
    request_id: filters.request_id?.trim(),
    from: toApiDateTime(filters.from),
    to: toApiDateTime(filters.to)
  };
}

export function toAdminAuditLogQueryString(filters: AuditLogFilters): string {
  return buildQueryString(toApiFilterPayload(filters));
}

export async function listAdminAuditLogs(filters: AuditLogFilters): Promise<AuditLogListResponse> {
  const query = toAdminAuditLogQueryString(filters);
  const response = await apiFetch<RawAuditLogListResponse>(appendQuery(AUDIT_LOGS_PATH, query));
  const logs = response.logs ?? response.audit_logs ?? response.items ?? [];

  return {
    logs: logs.map(normalizeAuditLog),
    total: numberValue(response.total),
    page: numberValue(response.page) ?? filters.page,
    page_size: numberValue(response.page_size ?? response.limit) ?? filters.page_size
  };
}

export async function exportAdminAuditLogs(request: AuditExportRequest): Promise<Blob> {
  return apiFetchBlob(AUDIT_LOG_EXPORT_PATH, {
    method: "POST",
    body: JSON.stringify({
      filters: toApiFilterPayload(request.filters),
      reason: request.reason.trim()
    })
  });
}
