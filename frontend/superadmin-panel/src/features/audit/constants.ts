import { AUDIT_RESOURCE_TYPES } from "./types";

export { AUDIT_RESOURCE_TYPES };

export const AUDIT_PAGE_SIZE_OPTIONS = [25, 50, 100] as const;
export const DEFAULT_AUDIT_PAGE_SIZE = 25;

export const HIGH_RISK_AUDIT_ACTIONS = [
  "user.blocked",
  "user.status.updated",
  "seller.suspended",
  "seller.status.updated",
  "refund.approved",
  "refund.rejected",
  "platform_setting.updated",
  "search_synonym.created",
  "search_synonym.updated",
  "search_synonym.deleted",
  "audit_logs.exported",
  "admin_data.exported"
] as const;

export const AUDIT_EXPORT_FILENAME_PREFIX = "audit-logs";
