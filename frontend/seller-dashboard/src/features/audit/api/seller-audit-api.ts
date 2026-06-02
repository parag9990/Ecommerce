import { http } from "../../../lib/http";
import type {
  AuditJson,
  AuditListParams,
  AuditPagination,
  SellerAuditLog,
  SellerAuditLogResponse,
} from "../types";
import { DEFAULT_AUDIT_PAGE_SIZE } from "../types";

type RawAuditResponse = {
  logs?: unknown[];
  audit_logs?: unknown[];
  items?: unknown[];
  pagination?: {
    page?: unknown;
    page_size?: unknown;
    total?: unknown;
    has_next?: unknown;
  };
  page?: unknown;
  page_size?: unknown;
  total?: unknown;
  has_next?: unknown;
};

function asRecord(value: unknown): Record<string, unknown> | null {
  return value && typeof value === "object" && !Array.isArray(value)
    ? (value as Record<string, unknown>)
    : null;
}

function stringOrEmpty(value: unknown) {
  return typeof value === "string" ? value : "";
}

function stringOrUndefined(value: unknown) {
  return typeof value === "string" && value.trim() !== "" ? value : undefined;
}

function positiveIntegerOr(value: unknown, fallback: number) {
  const parsed = Number(value);
  return Number.isInteger(parsed) && parsed > 0 ? parsed : fallback;
}

function optionalPositiveInteger(value: unknown) {
  const parsed = Number(value);
  return Number.isInteger(parsed) && parsed >= 0 ? parsed : undefined;
}

function auditJsonFrom(value: unknown): AuditJson {
  if (typeof value === "string" && value.trim() !== "") {
    try {
      return auditJsonFrom(JSON.parse(value) as unknown);
    } catch {
      return null;
    }
  }

  return asRecord(value);
}

export function normalizeSellerAuditLog(value: unknown): SellerAuditLog | null {
  const candidate = asRecord(value);

  if (!candidate) {
    return null;
  }

  if (
    typeof candidate.audit_id !== "string" ||
    typeof candidate.actor_user_id !== "string" ||
    typeof candidate.action !== "string" ||
    typeof candidate.resource_type !== "string" ||
    typeof candidate.resource_id !== "string" ||
    typeof candidate.created_at !== "string"
  ) {
    return null;
  }

  return {
    audit_id: candidate.audit_id,
    seller_id: stringOrEmpty(candidate.seller_id),
    actor_user_id: candidate.actor_user_id,
    actor_name: stringOrUndefined(candidate.actor_name),
    actor_email: stringOrUndefined(candidate.actor_email),
    action: candidate.action,
    resource_type: candidate.resource_type,
    resource_id: candidate.resource_id,
    resource_title: stringOrUndefined(candidate.resource_title),
    before: auditJsonFrom(candidate.before ?? candidate.before_json),
    after: auditJsonFrom(candidate.after ?? candidate.after_json),
    created_at: candidate.created_at,
  };
}

function normalizePagination(
  response: RawAuditResponse,
  logCount: number,
): AuditPagination {
  const pagination = response.pagination ?? {};
  const page = positiveIntegerOr(pagination.page ?? response.page, 1);
  const pageSize = positiveIntegerOr(
    pagination.page_size ?? response.page_size,
    DEFAULT_AUDIT_PAGE_SIZE,
  );
  const total = optionalPositiveInteger(pagination.total ?? response.total);
  const explicitHasNext = pagination.has_next ?? response.has_next;

  return {
    page,
    page_size: pageSize,
    total,
    has_next:
      typeof explicitHasNext === "boolean"
        ? explicitHasNext
        : typeof total === "number"
          ? page * pageSize < total
          : logCount >= pageSize,
  };
}

export function normalizeSellerAuditResponse(value: unknown): SellerAuditLogResponse {
  const response = asRecord(value) as RawAuditResponse | null;

  if (!response) {
    return {
      logs: [],
      pagination: {
        page: 1,
        page_size: DEFAULT_AUDIT_PAGE_SIZE,
        total: 0,
        has_next: false,
      },
    };
  }

  const rawLogs = response.logs ?? response.audit_logs ?? response.items ?? [];
  const logs = Array.isArray(rawLogs)
    ? rawLogs
        .map(normalizeSellerAuditLog)
        .filter((log): log is SellerAuditLog => log !== null)
    : [];

  return {
    logs,
    pagination: normalizePagination(response, logs.length),
  };
}

function setStringParam(search: URLSearchParams, key: string, value: unknown) {
  if (typeof value === "string" && value.trim() !== "") {
    search.set(key, value.trim());
  }
}

function toQuery(params: AuditListParams) {
  const search = new URLSearchParams();

  if (params.page && params.page > 0) {
    search.set("page", String(params.page));
  }

  if (params.page_size && params.page_size > 0) {
    search.set("page_size", String(params.page_size));
  }

  setStringParam(search, "actor_id", params.actor_id);
  setStringParam(search, "action", params.action);
  setStringParam(search, "resource_type", params.resource_type);
  setStringParam(search, "resource_id", params.resource_id);
  setStringParam(search, "from", params.from);
  setStringParam(search, "to", params.to);

  const query = search.toString();
  return query ? `?${query}` : "";
}

export async function listSellerAuditLogs(
  params: AuditListParams = {},
): Promise<SellerAuditLogResponse> {
  const response = await http<unknown>(`/api/v1/seller/audit-logs${toQuery(params)}`);

  return normalizeSellerAuditResponse(response);
}
