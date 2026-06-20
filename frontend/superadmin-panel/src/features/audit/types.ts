import type { AdminRole } from "../../lib/admin-rbac";

export type { AdminRole };

export const AUDIT_RESOURCE_TYPES = [
  "user",
  "seller",
  "order",
  "payment",
  "refund",
  "session",
  "platform_setting",
  "search_synonym",
  "admin_user",
  "admin_audit_logs"
] as const;

export type AuditResourceType = (typeof AUDIT_RESOURCE_TYPES)[number];

export type AuditSummary = Record<string, unknown>;

export type AdminAuditLog = {
  id: string;
  actor_admin_id: string;
  actor_role: AdminRole | string;
  action: string;
  resource_type: AuditResourceType | string;
  resource_id: string;
  request_id: string;
  ip_hash: string;
  before_summary: AuditSummary | null;
  after_summary: AuditSummary | null;
  reason: string | null;
  created_at: string | null;
};

export type AuditLogFilters = {
  actor_id?: string;
  action?: string;
  resource_type?: string;
  resource_id?: string;
  request_id?: string;
  from?: string;
  to?: string;
  page: number;
  page_size: number;
};

export type AuditLogListResponse = {
  logs: AdminAuditLog[];
  total?: number;
  page: number;
  page_size: number;
};

export type AuditExportRequest = {
  filters: AuditLogFilters;
  reason: string;
};
